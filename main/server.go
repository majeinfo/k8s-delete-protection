package main

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	admission "k8s.io/api/admission/v1beta1"
	k8meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
)

type AdmissionHandler struct {
}

// Handle requests
func (handler *AdmissionHandler) handler(w http.ResponseWriter, r *http.Request) {
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
	log.Debugf("AdmissionReview %s, %s, %s, %v, %v", request.Request.Name, request.Request.Namespace,
		request.Request.Operation, request.Request.Kind, request.Request.Resource)

	result, err := checkRequest(request.Request, handler)
	response := admission.AdmissionResponse{
		UID:     request.Request.UID,
		Allowed: result,
	}
	if err != nil {
		response.Result = &k8meta.Status{
			Message: fmt.Sprintf("%v", err),
			Reason:  k8meta.StatusReasonUnauthorized,
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
		log.Debugf("Failed reason: %v", outReview.Response.Result.Status)
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

func checkRequest(request *admission.AdmissionRequest, handler *AdmissionHandler) (bool, error) {
	// Sanity checks
	if request.Operation != "DELETE" {
		log.Infof("Skipped resource [%v,%v,%v], check rules to exclude this resource", request.RequestKind.Group, request.RequestKind.Kind, request.Operation)
		return true, nil
	}

	if request.Kind.Kind == "Pod" {
		return false, fmt.Errorf("cannot delete Pod now !")
	}

	return true, nil
}

