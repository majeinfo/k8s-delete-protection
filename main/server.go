package main

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	admission "k8s.io/api/admission/v1"
	k8meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
)

// Pod livenessProbe
func handleLiveness(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	w.WriteHeader(http.StatusOK)
}

// Handle requests
func handleAdmissionRequest(w http.ResponseWriter, r *http.Request) {
	var body []byte
	if r.Body != nil {
		data, err := ioutil.ReadAll(r.Body)
		if err == nil {
			body = data
		} else {
			log.Infof("Error %v", err)
			http.Error(w, "Error reading body", http.StatusBadRequest)
			return
		}
	}
	if len(body) == 0 {
		log.Error("Body is empty")
		http.Error(w, "Body is empty", http.StatusBadRequest)
		return
	}

	request := admission.AdmissionReview{}
	if err := json.Unmarshal(body, &request); err != nil {
		log.Errorf("Error parsing body %v", err)
		http.Error(w, "Error parsing body", http.StatusBadRequest)
		return
	}
	log.Debugf("AdmissionReview op=%s on %s %s/%s",
		request.Request.Operation,
		request.Request.Kind.Kind,
		request.Request.Namespace,
		request.Request.Name)

	result, err := checkRequest(request.Request)
	response := admission.AdmissionResponse{
		UID:     request.Request.UID,
		Allowed: result,
	}
	if err != nil {
		response.Result = &k8meta.Status{
			Message: fmt.Sprintf("%v", err),
			Reason:  k8meta.StatusReasonForbidden,
		}
	}

	outReview := admission.AdmissionReview{
		TypeMeta: request.TypeMeta,
		Request:  request.Request,
		Response: &response,
	}
	json, err := json.Marshal(outReview)
	log.Debugf("AdmissionResponse %v", outReview.Response.Allowed)
	if !outReview.Response.Allowed {
		log.Debugf("Failed reason: %v", outReview.Response.Result)
	}

	if err != nil {
		log.Errorf("json.Marshal error %v", err)
		http.Error(w, fmt.Sprintf("Error encoding response %v", err), http.StatusInternalServerError)
	} else {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(json); err != nil {
			log.Errorf("Error writing response %v", err)
			http.Error(w, fmt.Sprintf("Error writing response: %v", err), http.StatusInternalServerError)
		}
	}
}

//func checkRequest(request *admission.AdmissionRequest, handler *AdmissionHandler) (bool, error) {
func checkRequest(request *admission.AdmissionRequest) (bool, error) {
	// Sanity checks
	if request.Operation != "DELETE" {
		log.Infof("Skipped resource [%v,%v,%v], check rules to exclude this resource", request.RequestKind.Group, request.RequestKind.Kind, request.Operation)
		return true, nil
	}

	// Apply "must" rules
	for _, rule := range must_rules {
		if doesRuleApply(&rule, request) {
			// The rule.label must exist !
			log.Debugf("'must' rules match")
			if labels, err := getObjectLabels(request); err == nil {
				if _, present := labels[rule.Label]; !present {
					log.Errorf("Object must not be deleted because it does not have this label: %s", rule.Label)
					return false, fmt.Errorf("Object must not be deleted because it does not have this label: %s", rule.Label)
				}
			}
		}
	}

	// Apply "must-not" rules
	for _, rule := range must_not_rules {
		if doesRuleApply(&rule, request) {
			// The rule.label must not exist !
			log.Debugf("'must-not' rules match")
			if labels, err := getObjectLabels(request); err == nil {
				if _, present := labels[rule.Label]; present {
					log.Errorf("Object must not be deleted because it has this label: %s", rule.Label)
					return false, fmt.Errorf("Object must not be deleted because it has this label: %s", rule.Label)
				}
			}
		}
	}

	return true, nil
}

func doesRuleApply(rule *Rule, request *admission.AdmissionRequest) bool {
	// Rule syntax:
	// namespace: default
	// kinds:
	//   - pods
	//   - nodes
	// label: allowed-for-deletion

	// namespace must match
	if rule.Namespace != "*" && (rule.Namespace != request.Namespace) {
		log.Debugf("Namespace mismatch: rule: %s, request: %s", rule.Namespace, request.Namespace)
		return false
	}

	// kind must match
	match := false
	for _, kind := range rule.Kinds {
		if kind == "*" || kind == request.Kind.Kind {
			match = true
			break
		}
	}
	if !match {
		log.Debugf("Kinds mismatch: rule: %v, request: %s", rule.Kinds, request.Kind.Kind)
		return false
	}

	return true
}

func getObjectLabels(request *admission.AdmissionRequest) (map[string]string, error) {
	var result map[string]interface{}
	var metadata map[string]interface{}
	labels := make(map[string]string)

	// Try to get the object label without taking care of the object type (Pod, Node, ...)
	if err := json.Unmarshal(request.OldObject.Raw, &result); err != nil {
		log.Errorf("Could not unmarshal raw object: %v", err)
		return labels, err
	}
	log.Debugf("result=%v", result)
	metadata = result["metadata"].(map[string]interface{})
	log.Debugf("metadata=%v", metadata)
	for key, value := range metadata["labels"].(map[string]interface{}) {
		labels[key] = value.(string)
	}
	log.Debugf("labels=%v", labels)

	return labels, nil
}

