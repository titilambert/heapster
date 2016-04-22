// Copyright 2015 Google Inc. All Rights Reserved.
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

package json

import (
	"encoding/json"
//	"fmt"
	"net/url"
//	"strings"
	"sync"
	"time"

	json_common "k8s.io/heapster/common/json"
	"k8s.io/heapster/events/core"
//	metrics_core "k8s.io/heapster/metrics/core"
	kube_api "k8s.io/kubernetes/pkg/api"

	"github.com/golang/glog"
)

type jsonSink struct {
	client json_common.JsonClient
	sync.RWMutex
	c        json_common.JsonConfig
	dbExists bool
}

const (
	eventMeasurementName = "log/events"
	// Event special tags
	eventUID = "uid"
	// Value Field name
	valueField = "value"
	// Event special tags
	dbNotFoundError = "database not found"

	// Maximum number of json Points to be sent in one batch.
	maxSendBatchSize = 1000
)

func (sink *jsonSink) resetConnection() {
	glog.Infof("Json connection reset")
	sink.dbExists = false
	sink.client = nil
}

// Generate point value for event
func getEventValue(event *kube_api.Event) (string, error) {
	// TODO: check whether indenting is required.
	bytes, err := json.MarshalIndent(event, "", " ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

/*
func eventToPoint(event *kube_api.Event) (*json.Point, error) {
	value, err := getEventValue(event)
	if err != nil {
		return nil, err
	}

	point := json.Point{
		Measurement: eventMeasurementName,
		Time:        event.LastTimestamp.Time.UTC(),
		Fields: map[string]interface{}{
			valueField: value,
		},
		Tags: map[string]string{
			eventUID: string(event.UID),
		},
	}
	if event.InvolvedObject.Kind == "Pod" {
		point.Tags[metrics_core.LabelPodId.Key] = string(event.InvolvedObject.UID)
		point.Tags[metrics_core.LabelPodName.Key] = event.InvolvedObject.Name
	}
	point.Tags[metrics_core.LabelHostname.Key] = event.Source.Host
	return &point, nil
}
*/

func (sink *jsonSink) ExportEvents(eventBatch *core.EventBatch) {
	sink.Lock()
	defer sink.Unlock()

    event_counter := 0
    start := time.Now()
	for _, event := range eventBatch.Events {
        json_event, err := json.Marshal(event)
        if err != nil {
            glog.Warningf("Failed to convert event to point: %v", err)
        }
        _, err = sink.client.WriteString(string(json_event))
		if err != nil {
		    glog.Errorf("Failed to write: %v", err)
		}
        err = sink.client.Flush()
		if err != nil {
		    glog.Errorf("Failed to flush: %v", err)
		} else {
            event_counter++
        }
	}
    end := time.Now()
	glog.V(4).Infof("Exported %d data to influxDB in %s", event_counter, end.Sub(start))
}


func (sink *jsonSink) Name() string {
	return "Json Sink"
}

func (sink *jsonSink) Stop() {
	// nothing needs to be done.
}



// Returns a thread-safe implementation of core.EventSink for Json.
func new(c json_common.JsonConfig) core.EventSink {
	client, err := json_common.NewClient(c)
	if err != nil {
		glog.Errorf("issues while creating an Json sink: %v, will retry on use", err)
	}
	return &jsonSink{
		client: client, // can be nil
		c:      c,
	}
}

func CreateJsonSink(uri *url.URL) (core.EventSink, error) {
	config, err := json_common.BuildConfig(uri)
	if err != nil {
		return nil, err
	}
	sink := new(*config)
	glog.Infof("created json sink with options: path:%s", config.Path)
	return sink, nil
}
