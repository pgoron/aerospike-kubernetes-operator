module github.com/aerospike/aerospike-kubernetes-operator

go 1.13

require (
	github.com/aerospike/aerospike-management-lib v0.0.0-20201112142459-9dbbb66be120
	github.com/ashishshinde/aerospike-client-go v3.0.4-0.20200924015406-d85b25081637+incompatible
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/emicklei/go-restful v2.11.1+incompatible // indirect
	github.com/evanphx/json-patch v4.5.0+incompatible
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32 // indirect
	github.com/go-openapi/spec v0.19.4
	github.com/inconshreveable/log15 v0.0.0-20201112154412-8562bdadbbac
	github.com/jinzhu/copier v0.2.3 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/operator-framework/operator-sdk v0.18.0
	github.com/stretchr/testify v1.5.1
	github.com/tmc/scp v0.0.0-20170824174625-f7b48647feef // indirect
	github.com/yuin/gopher-lua v0.0.0-20200816102855-ee81675732da // indirect
	golang.org/x/crypto v0.0.0-20200414173820-0848c9571904
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/kube-openapi v0.0.0-20200121204235-bf4fb3bd569c
	sigs.k8s.io/controller-runtime v0.6.0
)

// // Pinned to kubernetes-1.16.2
// replace (
// 	k8s.io/api => k8s.io/api v0.0.0-20191016110408-35e52d86657a
// 	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20191016113550-5357c4baaf65
// 	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20191004115801-a2eda9f80ab8
// 	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20191016112112-5190913f932d
// 	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20191016114015-74ad18325ed5
// 	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20191016115326-20453efc2458
// 	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.0.0-20191016115129-c07a134afb42
// 	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20191004115455-8e001e5d1894
// 	k8s.io/component-base => k8s.io/component-base v0.0.0-20191016111319-039242c015a9
// 	k8s.io/cri-api => k8s.io/cri-api v0.0.0-20190828162817-608eb1dad4ac
// 	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.0.0-20191016115521-756ffa5af0bd
// 	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20191016112429-9587704a8ad4
// 	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.0.0-20191016114939-2b2b218dc1df
// 	k8s.io/kube-proxy => k8s.io/kube-proxy v0.0.0-20191016114407-2e83b6f20229
// 	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.0.0-20191016114748-65049c67a58b
// 	k8s.io/kubectl => k8s.io/kubectl v0.0.0-20191016120415-2ed914427d51
// 	k8s.io/kubelet => k8s.io/kubelet v0.0.0-20191016114556-7841ed97f1b2
// 	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.0.0-20191016115753-cf0698c3a16b
// 	k8s.io/metrics => k8s.io/metrics v0.0.0-20191016113814-3b1a734dba6e
// 	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.0.0-20191016112829-06bb3c9d77c9
// )

replace github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309 // Required by Helm

replace github.com/openshift/api => github.com/openshift/api v0.0.0-20190924102528-32369d4db2ad // Required until https://github.com/operator-framework/operator-lifecycle-manager/pull/1241 is resolved

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
	k8s.io/client-go => k8s.io/client-go v0.18.2 // Required by prometheus-operator
)

// require (
// 	github.com/aerospike/aerospike-management-lib v0.0.0-20201112142459-9dbbb66be120
// 	github.com/ashishshinde/aerospike-client-go v3.0.4-0.20200924015406-d85b25081637+incompatible
// 	github.com/elazarl/goproxy v0.0.0-20190421051319-9d40249d3c2f // indirect
// 	github.com/elazarl/goproxy/ext v0.0.0-20190421051319-9d40249d3c2f // indirect
// 	github.com/evanphx/json-patch v4.9.0+incompatible
// 	github.com/go-openapi/spec v0.19.4
// 	github.com/inconshreveable/log15 v0.0.0-20180818164646-67afb5ed74ec
// 	github.com/jinzhu/copier v0.0.0-20190924061706-b57f9002281a // indirect
// 	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
// 	github.com/operator-framework/operator-sdk v0.17.0
// 	github.com/spf13/cobra v1.0.0 // indirect
// 	github.com/stretchr/testify v1.5.1
// 	github.com/tmc/scp v0.0.0-20170824174625-f7b48647feef // indirect
// 	github.com/travelaudience/aerospike-operator v0.0.0-20191002090530-354c1a4e7e2a
// 	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
// 	golang.org/x/crypto v0.0.0-20200220183623-bac4c82f6975
// 	golang.org/x/tools v0.0.0-20200616195046-dc31b401abb5 // indirect
// 	k8s.io/api v0.20.1
// 	k8s.io/apimachinery v0.20.1
// 	k8s.io/client-go v12.0.0+incompatible
// 	k8s.io/kube-openapi v0.0.0-20191107075043-30be4d16710a
// 	sigs.k8s.io/controller-runtime v0.5.2
// )

// require github.com/Azure/go-autorest v12.2.0+incompatible // indirect

// require (
// 	github.com/MakeNowJust/heredoc v0.0.0-20171113091838-e9091a26100e // indirect
// 	github.com/ant31/crd-validation v0.0.0-20180702145049-30f8a35d0ac2 // indirect
// 	github.com/bugsnag/bugsnag-go v1.5.0 // indirect
// 	github.com/bugsnag/panicwrap v1.2.0 // indirect
// 	github.com/cloudflare/cfssl v0.0.0-20180726162950-56268a613adf // indirect
// 	github.com/coreos/go-systemd v0.0.0-20190620071333-e64a0ec8b42a // indirect
// 	github.com/coreos/rkt v1.30.0 // indirect
// 	github.com/datastax/cass-operator v1.5.0 // indirect
// 	github.com/docker/go-metrics v0.0.0-20181218153428-b84716841b82 // indirect
// 	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7 // indirect
// 	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
// 	github.com/emicklei/go-restful v2.11.1+incompatible // indirect
// 	github.com/garyburd/redigo v1.6.0 // indirect
// 	github.com/go-logr/logr v0.3.0 // indirect
// 	github.com/go-logr/zapr v0.2.0 // indirect
// 	github.com/godbus/dbus v4.1.0+incompatible // indirect
// 	github.com/gofrs/uuid v3.2.0+incompatible // indirect
// 	github.com/golang/lint v0.0.0-20180702182130-06c8688daad7 // indirect
// 	github.com/google/certificate-transparency-go v1.0.21 // indirect
// 	github.com/googleapis/gnostic v0.4.0 // indirect
// 	github.com/gorilla/handlers v1.4.0 // indirect
// 	github.com/gregjones/httpcache v0.0.0-20190203031600-7a902570cb17 // indirect
// 	github.com/hashicorp/golang-lru v0.5.4 // indirect
// 	github.com/heketi/rest v0.0.0-20180404230133-aa6a65207413 // indirect
// 	github.com/heketi/utils v0.0.0-20170317161834-435bc5bdfa64 // indirect
// 	github.com/imdario/mergo v0.3.10 // indirect
// 	github.com/improbable-eng/thanos v0.3.2 // indirect
// 	github.com/kr/pretty v0.2.0 // indirect
// 	github.com/lib/pq v1.2.0 // indirect
// 	github.com/mattbaird/jsonpatch v0.0.0-20171005235357-81af80346b1a // indirect
// 	github.com/onsi/ginkgo v1.14.1 // indirect
// 	github.com/onsi/gomega v1.10.2 // indirect
// 	github.com/openshift/api v3.9.1-0.20190924102528-32369d4db2ad+incompatible // indirect
// 	github.com/prometheus/client_golang v1.7.1 // indirect
// 	github.com/prometheus/tsdb v0.8.0 // indirect
// 	github.com/xenolf/lego v0.3.2-0.20160613233155-a9d8cec0e656 // indirect
// 	github.com/xlab/handysort v0.0.0-20150421192137-fb3537ed64a1 // indirect
// 	github.com/yvasiyarov/go-metrics v0.0.0-20150112132944-c25f46c4b940 // indirect
// 	github.com/yvasiyarov/gorelic v0.0.6 // indirect
// 	go.uber.org/zap v1.15.0 // indirect
// 	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e // indirect
// 	gomodules.xyz/jsonpatch/v2 v2.1.0 // indirect
// 	gonum.org/v1/gonum v0.0.0-20190710053202-4340aa3071a0 // indirect
// 	google.golang.org/appengine v1.6.6 // indirect
// 	gopkg.in/square/go-jose.v1 v1.1.2 // indirect
// 	k8s.io/apiextensions-apiserver v0.20.1 // indirect
// 	k8s.io/utils v0.0.0-20201110183641-67b214c5f920 // indirect
// 	sigs.k8s.io/testing_frameworks v0.1.2 // indirect
// 	vbom.ml/util v0.0.0-20160121211510-db5cfe13f5cc // indirect
// )

// // Pinned to kubernetes-1.16.2
// replace (
// 	k8s.io/api => k8s.io/api v0.0.0-20191016110408-35e52d86657a
// 	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20191016113550-5357c4baaf65
// 	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20191004115801-a2eda9f80ab8
// 	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20191016112112-5190913f932d
// 	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20191016114015-74ad18325ed5
// 	k8s.io/client-go => k8s.io/client-go v0.0.0-20191016111102-bec269661e48
// 	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20191016115326-20453efc2458
// 	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.0.0-20191016115129-c07a134afb42
// 	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20191004115455-8e001e5d1894
// 	k8s.io/component-base => k8s.io/component-base v0.0.0-20191016111319-039242c015a9
// 	k8s.io/cri-api => k8s.io/cri-api v0.0.0-20190828162817-608eb1dad4ac
// 	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.0.0-20191016115521-756ffa5af0bd
// 	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20191016112429-9587704a8ad4
// 	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.0.0-20191016114939-2b2b218dc1df
// 	k8s.io/kube-proxy => k8s.io/kube-proxy v0.0.0-20191016114407-2e83b6f20229
// 	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.0.0-20191016114748-65049c67a58b
// 	k8s.io/kubectl => k8s.io/kubectl v0.0.0-20191016120415-2ed914427d51
// 	k8s.io/kubelet => k8s.io/kubelet v0.0.0-20191016114556-7841ed97f1b2
// 	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.0.0-20191016115753-cf0698c3a16b
// 	k8s.io/metrics => k8s.io/metrics v0.0.0-20191016113814-3b1a734dba6e
// 	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.0.0-20191016112829-06bb3c9d77c9
// )

// replace github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309 // Required by Helm

// replace github.com/openshift/api => github.com/openshift/api v0.0.0-20190924102528-32369d4db2ad // Required until https://github.com/operator-framework/operator-lifecycle-manager/pull/1241 is resolved
