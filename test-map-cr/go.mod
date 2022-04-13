module test-map-cr

go 1.17

require (
	github.com/hazelcast/hazelcast-go-client v1.3.0
	github.com/hazelcast/hazelcast-platform-operator v0.2.0
)

require (
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/go-logr/logr v0.3.0 // indirect
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/shirou/gopsutil/v3 v3.21.5 // indirect
	github.com/tklauser/go-sysconf v0.3.4 // indirect
	github.com/tklauser/numcpus v0.2.1 // indirect
	golang.org/x/net v0.0.0-20210805182204-aaa1db679c0d // indirect
	golang.org/x/sys v0.0.0-20211216021012-1d35b9e2eb4e // indirect
	golang.org/x/text v0.3.6 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/api v0.20.2 // indirect
	k8s.io/apimachinery v0.20.2 // indirect
	k8s.io/klog/v2 v2.4.0 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.0.2 // indirect
)

replace github.com/hazelcast/hazelcast-go-client v1.3.0 => github.com/hazelcast/hazelcast-go-client v1.1.2-0.20220407121529-67a54909b377

replace github.com/hazelcast/hazelcast-platform-operator v0.2.0 => /home/mty/Whz/hazelcast-enterprise-operator
