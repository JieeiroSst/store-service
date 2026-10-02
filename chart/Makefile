k8s_generate_name_prev:
	helm install ../chart --dry-run --generate-name
k8s_generate_name:
	helm install ../chart --generate-name
# Charts on the common library carry their own file:// dependency, which the
# umbrella does not build recursively, so build those first.
k8s_subchart_deps:
	for d in $$(grep -l 'file://../common' */Chart.yaml | xargs -n1 dirname); do helm dependency build $$d; done
k8s_build: k8s_subchart_deps
	helm dependency build  
k8s_update: k8s_subchart_deps
	helm dependency update 
k8s_list:
	helm list
k8s_remove:
	helm delete chart-*
log:
	kubectl logs ${pod} --all-containers
upgrade:
	helm upgrade $(version) ../chart
# port-forward:
#     kubectl port-forward service/${service} ${pod}
