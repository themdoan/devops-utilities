/*
Copyright 2020 The Flux authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cloudtrace

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	cloudtracepb "cloud.google.com/go/trace/apiv1/tracepb"
)

type Alertmanager struct {
	URL      string
	ProxyURL string
	CertPool *x509.CertPool
}

type AlertManagerAlert struct {
	Status      string            `json:"status"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`

	StartsAt AlertManagerTime `json:"startsAt"`
	EndsAt   AlertManagerTime `json:"endsAt,omitempty"`
}

// AlertManagerTime takes care of representing time.Time as RFC3339.
// See https://prometheus.io/docs/alerting/0.27/clients/
type AlertManagerTime time.Time

func (a AlertManagerTime) String() string {
	return time.Time(a).Format(time.RFC3339)
}

func (a AlertManagerTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

func (a *AlertManagerTime) UnmarshalJSON(jsonRepr []byte) error {
	var serializedTime string
	if err := json.Unmarshal(jsonRepr, &serializedTime); err != nil {
		return err
	}

	t, err := time.Parse(time.RFC3339, serializedTime)
	if err != nil {
		return err
	}

	*a = AlertManagerTime(t)
	return nil
}

func NewAlertmanager(hookURL string, proxyURL string, certPool *x509.CertPool) (*Alertmanager, error) {
	_, err := url.ParseRequestURI(hookURL)
	if err != nil {
		return nil, fmt.Errorf("invalid Alertmanager URL %s: '%w'", hookURL, err)
	}

	return &Alertmanager{
		URL:      hookURL,
		ProxyURL: proxyURL,
		CertPool: certPool,
	}, nil
}

func (s *Alertmanager) Post(ctx context.Context, entries []*cloudtracepb.Trace) error {

	payload := []*AlertManagerAlert{}

	for _, trace := range entries {
		var alertMessage AlertManagerAlert
		alertMessage.Labels = make(map[string]string)
		alertMessage.Annotations = make(map[string]string)

		alertMessage.Status = "firing"
		alertMessage.Annotations["description"] = "Query slower than 5s"
		alertMessage.Labels = trace.Spans[0].Labels
		alertMessage.Labels["alertname"] = "Slow Query"

		payload = append(payload, &alertMessage)

	}

	// fmt.Printf("payload: %v", payload)
	err := postMessage(ctx, s.URL, s.ProxyURL, s.CertPool, payload)

	if err != nil {
		return fmt.Errorf("postMessage failed: %w", err)
	}
	return nil
}
