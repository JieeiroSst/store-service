k8s_generate_name_prev:
	helm install ../chart-k8s --dry-run --generate-name
k8s_generate_name:
	helm install ../chart-k8s --generate-name
k8s_build:
	helm dependency build  
k8s_update:
	helm dependency update 
k8s_list:
	helm list
k8s_remove:
	helm delete chart-*
log:
	kubectl logs ${pod} --all-containers
upgrade:
	helm upgrade $(version) ../chart-k8s
# port-forward:
#     kubectl port-forward service/${service} ${pod}