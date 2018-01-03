// Copyright 2017 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package heapster

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"errors"

	"github.com/kubernetes/dashboard/src/app/backend/api"
	"github.com/kubernetes/dashboard/src/app/backend/client"
	integrationapi "github.com/kubernetes/dashboard/src/app/backend/integration/api"
	metricapi "github.com/kubernetes/dashboard/src/app/backend/integration/metric/api"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	heapster "k8s.io/heapster/metrics/api/v1/types"
)

func areErrorsEqual(err1, err2 error) bool {
	return (err1 != nil && err2 != nil && err1.Error() == err2.Error()) ||
		(err1 == nil && err2 == nil)
}

type GlobalCounter int32

func (c *GlobalCounter) increment() int32 {
	return atomic.AddInt32((*int32)(c), 1)
}

func (c *GlobalCounter) get() int32 {
	return atomic.LoadInt32((*int32)(c))
}

func (c *GlobalCounter) set(val int32) {
	atomic.StoreInt32((*int32)(c), val)
}

var _NumRequests = GlobalCounter(0)

type FakeHeapster struct {
	PodData
	NodeData
	numRequests int
}

type FakeRequest struct {
	PodData
	NodeData
	Path string
}

type PodData map[string][]heapster.MetricPoint
type NodeData map[string][]heapster.MetricPoint

func (self FakeHeapster) Get(path string) RequestInterface {
	return FakeRequest{self.PodData, self.NodeData, path}
}

func (self FakeHeapster) GetNumberOfRequestsMade() int {
	num := int(_NumRequests.get())
	_NumRequests.set(0)
	return num
}

func (self FakeHeapster) HealthCheck() error {
	return nil
}

func (self FakeHeapster) ID() integrationapi.IntegrationID {
	return "fakeHeapster"
}

func (self FakeRequest) DoRaw() ([]byte, error) {
	_NumRequests.increment()
	log.Println("Performing req...")
	path := self.Path
	time.Sleep(50 * time.Millisecond) // simulate response delay of 0.05 seconds
	if strings.Contains(path, "/pod-list/") {
		r, _ := regexp.Compile(`\/pod\-list\/(.+)\/metrics\/`)
		submatch := r.FindStringSubmatch(path)
		if len(submatch) != 2 {
			return nil, fmt.Errorf("Invalid request url %s", path)
		}
		requestedPods := strings.Split(submatch[1], ",")

		r, _ = regexp.Compile(`\/namespaces\/(.+)\/pod\-list\/`)
		submatch = r.FindStringSubmatch(path)
		if len(submatch) != 2 {
			return nil, fmt.Errorf("Invalid request url %s", path)
		}
		namespace := submatch[1]

		items := []heapster.MetricResult{}
		for _, pod := range requestedPods {
			items = append(items, heapster.MetricResult{Metrics: self.PodData[pod+"/"+namespace]})
		}
		x, err := json.Marshal(heapster.MetricResultList{Items: items})
		log.Println("Got you:", string(x))
		return x, err

	} else if strings.Contains(path, "/nodes/") {
		r, _ := regexp.Compile(`\/nodes\/(.+)\/metrics\/`)
		submatch := r.FindStringSubmatch(path)
		if len(submatch) != 2 {
			return nil, fmt.Errorf("Invalid request url %s", path)
		}
		requestedNode := submatch[1]

		x, err := json.Marshal(heapster.MetricResult{Metrics: self.NodeData[requestedNode]})
		log.Println("Got you:", string(x))
		return x, err
	} else {
		return nil, fmt.Errorf("Invalid request url %s", path)
	}
}

func (self FakeRequest) AbsPath(segments ...string) *rest.Request {
	return &rest.Request{}
}

const TimeTemplate = "2016-08-12T11:0%d:00Z"
const TimeTemplateValue = int64(1470999600)

func NewRawDPs(dps []int64, startTime int) []heapster.MetricPoint {
	newRdps := []heapster.MetricPoint{}
	for i := 0; i < len(dps) && startTime+i < 10; i++ {
		parsedTime, _ := time.Parse(time.RFC3339, fmt.Sprintf(TimeTemplate, i+startTime))
		newRdps = append(newRdps, heapster.MetricPoint{Timestamp: parsedTime, Value: uint64(dps[i])})
	}
	return newRdps
}

func newDps(dps []int64, startTime int) metricapi.DataPoints {
	newDps := metricapi.DataPoints{}
	for i := 0; i < len(dps) && startTime+i < 10; i++ {
		newDps = append(newDps, metricapi.DataPoint{TimeTemplateValue + int64(60*(i+startTime)), dps[i]})
	}
	return newDps
}

var fakePodData = PodData{
	"P1/a": NewRawDPs([]int64{0, 5, 10}, 0),
	"P2/a": NewRawDPs([]int64{15, 20, 25}, 0),
	"P3/a": NewRawDPs([]int64{30, 35, 40}, 0),
	"P4/a": NewRawDPs([]int64{45, 50, -100000}, 0),
	"P1/b": NewRawDPs([]int64{1000, 1100}, 0),
	"P2/b": NewRawDPs([]int64{1200, 1300}, 1),
	"P3/b": NewRawDPs([]int64{1400, 1500}, 2),
	"P4/b": NewRawDPs([]int64{}, 0),
	"P1/c": NewRawDPs([]int64{10000, 11000, 12000}, 0),
	"P2/c": NewRawDPs([]int64{13000, 14000, 15000}, 0),
}

var fakeNodeData = NodeData{
	"N1": NewRawDPs([]int64{0, 5, 10}, 0),
	"N2": NewRawDPs([]int64{15, 20, 25}, 0),
	"N3": NewRawDPs([]int64{30, 35, 40}, 0),
	"N4": NewRawDPs([]int64{45, 50, 55}, 0),
}

var fakeHeapsterClient = FakeHeapster{
	PodData:  fakePodData,
	NodeData: fakeNodeData,
}

func getResourceSelector(namespace string, resourceType api.ResourceKind,
	resourceName, uid string) metricapi.ResourceSelector {
	return metricapi.ResourceSelector{
		Namespace:    namespace,
		ResourceType: resourceType,
		ResourceName: resourceName,
		UID:          types.UID(uid),
	}
}

func TestDownloadMetric(t *testing.T) {
	type HeapsterSelectorTestCase struct {
		Info                string
		Selectors           []metricapi.ResourceSelector
		ExpectedDataPoints  metricapi.DataPoints
		ExpectedNumRequests int
	}
	testCases := []HeapsterSelectorTestCase{
		{
			"get data for single pod",
			[]metricapi.ResourceSelector{
				getResourceSelector("a", api.ResourceKindPod, "P1", "U1"),
			},
			newDps([]int64{0, 5, 10}, 0),
			1,
		},
		{
			"get data for 3 pods",
			[]metricapi.ResourceSelector{
				getResourceSelector("a", api.ResourceKindPod, "P1", "U1"),
				getResourceSelector("a", api.ResourceKindPod, "P2", "U2"),
				getResourceSelector("a", api.ResourceKindPod, "P3", "U3"),
			},
			newDps([]int64{45, 60, 75}, 0),
			1,
		},
		{
			"get data for 4 pods where 1 pod does not exist - ignore non existing pod",
			[]metricapi.ResourceSelector{
				getResourceSelector("a", api.ResourceKindPod, "P1", "U1"),
				getResourceSelector("a", api.ResourceKindPod, "P2", "U2"),
				getResourceSelector("a", api.ResourceKindPod, "P3", "U3"),
				getResourceSelector("a", api.ResourceKindPod, "NON_EXISTING", "NA"),
			},
			newDps([]int64{45, 60, 75}, 0),
			1,
		},
		{
			"get data for 4 pods where pods have different X timestams available",
			[]metricapi.ResourceSelector{
				getResourceSelector("b", api.ResourceKindPod, "P1", "U1"),
				getResourceSelector("b", api.ResourceKindPod, "P2", "U2"),
				getResourceSelector("b", api.ResourceKindPod, "P3", "U3"),
				getResourceSelector("b", api.ResourceKindPod, "P4", "U4"),
			},
			newDps([]int64{1000, 2300, 2700, 1500}, 0),
			1,
		},
		{
			"ask for non existing namespace - return no data points",
			[]metricapi.ResourceSelector{
				getResourceSelector("NON_EXISTING_NAMESPACE", api.ResourceKindPod,
					"P1", "U1"),
			},
			newDps([]int64{}, 0),
			1,
		},
		{
			"get data for 0 pods - return no data points",
			[]metricapi.ResourceSelector{},
			newDps([]int64{}, 0),
			0,
		},
		{
			"get data for 0 nodes - return no data points",
			[]metricapi.ResourceSelector{},
			newDps([]int64{}, 0),
			0,
		},
		{
			"ask for 1 node",
			[]metricapi.ResourceSelector{
				getResourceSelector("NO_NAMESPACE", api.ResourceKindNode, "N1",
					"U11"),
			},
			newDps([]int64{0, 5, 10}, 0),
			1,
		},
		{
			"ask for 3 nodes",
			[]metricapi.ResourceSelector{
				getResourceSelector("NO_NAMESPACE", api.ResourceKindNode, "N1",
					"U11"),
				getResourceSelector("NO_NAMESPACE", api.ResourceKindNode, "N2",
					"U12"),
				getResourceSelector("NO_NAMESPACE", api.ResourceKindNode, "N3",
					"U13"),
			},
			newDps([]int64{45, 60, 75}, 0),
			3, // change this to 1 when nodes support all in 1 download.
		},
	}
	for _, testCase := range testCases {
		log.Println("-----------\n\n\n", testCase.Info, int(_NumRequests.get()))
		hClient := heapsterClient{fakeHeapsterClient}
		promises := hClient.DownloadMetric(testCase.Selectors, "",
			&metricapi.CachedResources{})
		metrics, err := hClient.AggregateMetrics(promises, "", nil).GetMetrics()
		if err != nil {
			t.Errorf("Test Case: %s. Failed to get metrics - %s", testCase.Info, err)
			return
		}
		num_req := fakeHeapsterClient.GetNumberOfRequestsMade()

		if !reflect.DeepEqual(metrics[0].DataPoints, testCase.ExpectedDataPoints) {
			t.Errorf("Test Case: %s. Received incorrect data points. Got %v, expected %v.",
				testCase.Info, metrics[0].DataPoints, testCase.ExpectedDataPoints)
		}

		if testCase.ExpectedNumRequests != num_req {
			t.Errorf("Test Case: %s. Selector performed unexpected number of requests to the heapster server. Performed %d, expected %d",
				testCase.Info, num_req, testCase.ExpectedNumRequests)
		}
	}
}

var selectorPool = []metricapi.ResourceSelector{
	getResourceSelector("a", api.ResourceKindPod, "P1", "U1"),
	getResourceSelector("a", api.ResourceKindPod, "P2", "U2"),
	getResourceSelector("a", api.ResourceKindPod, "P3", "U3"),
	getResourceSelector("b", api.ResourceKindPod, "P1", "Z1"),
	getResourceSelector("b", api.ResourceKindPod, "P2", "Z2"),
	getResourceSelector("b", api.ResourceKindPod, "P3", "Z3"),
	getResourceSelector("NO_NAMESPACE", api.ResourceKindNode, "N1", "U11"),
	getResourceSelector("NO_NAMESPACE", api.ResourceKindNode, "N2", "U12"),
	getResourceSelector("NO_NAMESPACE", api.ResourceKindNode, "N3", "U13"),
	getResourceSelector("NO_NAMESPACE", api.ResourceKindNode, "N4", "U14"),
}

func TestDownloadMetrics(t *testing.T) {
	type HeapsterSelectorsTestCase struct {
		Info                string
		SelectorIds         []int
		AggregationNames    metricapi.AggregationModes
		MetricNames         []string
		ExpectedDataPoints  []metricapi.DataPoints
		ExpectedNumRequests int
	}

	MinMaxSumAggregations := metricapi.AggregationModes{metricapi.MinAggregation,
		metricapi.MaxAggregation, metricapi.SumAggregation}
	testCases := []HeapsterSelectorsTestCase{
		{
			"ask for 1 resource",
			[]int{1},
			MinMaxSumAggregations,
			[]string{"Dummy/Metric"},
			[]metricapi.DataPoints{
				newDps([]int64{15, 20, 25}, 0),
				newDps([]int64{15, 20, 25}, 0),
				newDps([]int64{15, 20, 25}, 0),
			},
			1,
		},
		{
			"ask for 2 resources from same namespace",
			[]int{0, 1},
			MinMaxSumAggregations,
			[]string{"Dummy/Metric"},
			[]metricapi.DataPoints{
				newDps([]int64{0, 5, 10}, 0),
				newDps([]int64{15, 20, 25}, 0),
				newDps([]int64{15, 25, 35}, 0),
			},
			1,
		},
		{
			"ask for 3 resources from same namespace, get 2 metrics",
			[]int{0, 1, 2},
			MinMaxSumAggregations,
			[]string{"Dummy/Metric1", "DummyMetric2"},
			[]metricapi.DataPoints{
				newDps([]int64{0, 5, 10}, 0),
				newDps([]int64{30, 35, 40}, 0),
				newDps([]int64{45, 60, 75}, 0),
				newDps([]int64{0, 5, 10}, 0),
				newDps([]int64{30, 35, 40}, 0),
				newDps([]int64{45, 60, 75}, 0),
			},
			2,
		},
		{
			"ask for multiple resources of the same kind from multiple namespaces",
			[]int{0, 1, 3, 4},
			MinMaxSumAggregations,
			[]string{"Dummy/Metric"},
			[]metricapi.DataPoints{
				newDps([]int64{0, 5, 10}, 0),
				newDps([]int64{1000, 1200, 1300}, 0),
				newDps([]int64{1015, 2325, 1335}, 0),
			},
			2,
		},
		{
			"ask for multiple resources of different kind from multiple namespaces",
			[]int{0, 1, 6, 7},
			MinMaxSumAggregations,
			[]string{"Dummy/Metric"},
			[]metricapi.DataPoints{
				newDps([]int64{0, 5, 10}, 0),
				newDps([]int64{15, 20, 25}, 0),
				newDps([]int64{30, 50, 70}, 0),
			},
			3, // if we had node-list option in heapster API we would make only 2
			// requests unfortunately there is no such option and we have to make one request per node
			// note that nodes overlap (1,2,3) + (3,4) and we download node 3 only once thanks to request compression
			// So 4 requests for nodes (one for each unique node) and 2 requests for pods (1 for each  namespace) = 6 in total.
		},
	}

	for _, testCase := range testCases {
		selectors := []metricapi.ResourceSelector{}
		hClient := heapsterClient{fakeHeapsterClient}
		for _, selectorId := range testCase.SelectorIds {
			selectors = append(selectors, selectorPool[selectorId])
		}

		metricPromises := make(metricapi.MetricPromises, 0)
		for _, metricName := range testCase.MetricNames {
			promises := hClient.DownloadMetric(selectors, metricName,
				&metricapi.CachedResources{})
			promises = hClient.AggregateMetrics(promises, metricName,
				testCase.AggregationNames)
			metricPromises = append(metricPromises, promises...)
		}
		metrics, err := metricPromises.GetMetrics()
		if err != nil {
			t.Errorf("Test Case: %s. Failed to get metrics - %s", testCase.Info, err)
			return
		}

		receivedDataPoints := []metricapi.DataPoints{}
		for _, metric := range metrics {
			receivedDataPoints = append(receivedDataPoints, metric.DataPoints)
		}

		if !reflect.DeepEqual(receivedDataPoints, testCase.ExpectedDataPoints) {
			t.Errorf("Test Case: %s. Received incorrect data points. Got %v, expected %v.",
				testCase.Info, receivedDataPoints, testCase.ExpectedDataPoints)
		}
		num_req := fakeHeapsterClient.GetNumberOfRequestsMade()
		if testCase.ExpectedNumRequests != num_req {
			t.Errorf("Test Case: %s. Selector performed unexpected number of requests to the heapster server. Performed %d, expected %d",
				testCase.Info, num_req, testCase.ExpectedNumRequests)
		}
	}
}

func TestCreateHeapsterClient(t *testing.T) {
	k8sClient := client.NewClientManager("", "http://localhost:8080").InsecureClient()
	cases := []struct {
		info         string
		heapsterHost string
		client       kubernetes.Interface
		expected     HeapsterRESTClient
		expectedErr  error
	}{
		{
			"should create in-cluster heapster client",
			"",
			k8sClient,
			inClusterHeapsterClient{},
			nil,
		},
		{
			"should create remote heapster client",
			"http://localhost:80801",
			nil,
			remoteHeapsterClient{},
			nil,
		},
		{
			"should return error",
			"invalid-url-!!23*%.",
			nil,
			nil,
			errors.New("parse http://invalid-url-!!23*%.: invalid URL escape \"%.\""),
		},
	}

	for _, c := range cases {
		metricClient, err := CreateHeapsterClient(c.heapsterHost, c.client)

		if !areErrorsEqual(c.expectedErr, err) {
			t.Errorf("Test Case: %s. Expected error to be: %v, but got %v.",
				c.info, c.expectedErr, err)
		}

		heapsterClient, _ := metricClient.(heapsterClient)
		if reflect.TypeOf(heapsterClient.client) != reflect.TypeOf(c.expected) {
			t.Errorf("Test Case: %s. Expected client to be of type: %v, but got %v",
				c.info, reflect.TypeOf(c.expected), reflect.TypeOf(heapsterClient.client))
		}
	}
}
