k8s_generate_name_prev:
	helm install ../chart --dry-run --generate-name
k8s_generate_name:
	helm install ../chart --generate-name
k8s_build:
	helm dependency build  
k8s_update:
	helm dependency update 
k8s_list:
	helm list
k8s_remove:
	helm delete chart-*
k8s_upgrade:
	helm upgrade 
log:
	kubectl logs ${pod} --all-containers
upgrade:
	helm upgrade $(version) ../chart
port-forward:
    kubectl port-forward service/${service} ${pod}
