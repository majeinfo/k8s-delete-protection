# k8s-delete-protection Admission Controller

If you want to make your Kubernetes cluster more robust and avoid to delete 
crucial resources, you may want to deploy this Admission Controller.

It is called "k8s-delete-protection" and can be easily configured to
prohibit deletion of resources.

Configuration rules may be "must" rules or "must-not" rules.

For example, if you call the Controller with the following "must" and "must-not" rules :

```
# must rules
- namespace: default
  kinds:
    - Pod
    - Deployment
  label: allowed-for-deletion
- namespace: "*"
  kinds:
    - Node
  label: allowed-for-deletion
```

```
# must-not rules
- namespace: "*"
  kinds:
    - "*"
  label: protected-against-deletion
```

Now, you cannot delete any node unless you label it like this :

`$ kubectl label <node> allowed-for-deletion=true`

The same rule applies to any Pod or Deployment belonging to the `default` namespace.

Furthermore, you cannot delete any resource from any namespace if it is labeled with `protected-against-deletion`.
You must unlabel it first :

`$ kubectl label <resource> protected-against-deletion-`

# Command line parameters

```
$ ./main [--cert certificate_filename (default: ./server.pem)]
         [--key private_key_filename (default: ./server-key.pem)]
         [--port listening_port (default: 8443)]
         [--must-rules filename (default: ./must.rules)]
         [--must-not-rules filename (default: ./must-not.rules)]
```

# Installation
## Object relationship
![Objects](https://user-images.githubusercontent.com/14954414/157268992-211a550c-8b82-4dac-b382-aca9f44b7bf6.jpg)

Once the key and certificate are generated, you can modify the rules-file content, then you apply the manifests in the following order :
```
- secrets-ca.yaml
- configmap.yaml
- deployment.yaml
- service.yaml
- webhook.yaml
```

A Helm Chart is coming soon...
