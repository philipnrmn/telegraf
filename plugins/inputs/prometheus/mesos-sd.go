package prometheus

import (
	"context"
	"io"
	"log"
	"net"
	"net/url"
	"strconv"

	"github.com/mesos/mesos-go/api/v1/lib/agent"
	"github.com/mesos/mesos-go/api/v1/lib/agent/calls"
	"github.com/mesos/mesos-go/api/v1/lib/httpcli"
	"github.com/mesos/mesos-go/api/v1/lib/httpcli/httpagent"
)

func discoverMesosTasks(mesosAgentUrl string) ([]URLAndAddress, error) {
	var result = make([]URLAndAddress, 0)

	// TODO error handling
	maUrl, _ := url.Parse(mesosAgentUrl)
	maHost, _, _ := net.SplitHostPort(maUrl.Host)

	// TODO: timeout
	uri := mesosAgentUrl + "/api/v1"
	cli := httpagent.NewSender(httpcli.New(httpcli.Endpoint(uri)).Send)
	ctx := context.Background()

	// TODO error handling
	resp, _ := cli.Send(ctx, calls.NonStreaming(calls.GetTasks()))

	defer func() {
		if resp != nil {
			resp.Close()
		}
	}()

	var r agent.Response
	for {
		if err := resp.Decode(&r); err != nil {
			if err == io.EOF {
				break
			}
			return result, err
		}
	}
	if r.GetType() == agent.Response_GET_TASKS {
		gt := r.GetGetTasks()
		log.Printf("prometheus::mesos: found %d tasks", len(gt.LaunchedTasks))
		for _, task := range gt.LaunchedTasks {
			log.Printf("prometheus::mesos: checking task named %s", task.Name)
			if task.Discovery == nil {
				continue
			}
			if task.Discovery.Ports == nil {
				continue
			}
			log.Printf("prometheus::mesos: %s had discovery and ports", task.Name)
			for _, port := range task.Discovery.Ports.Ports {
				if port.Labels == nil {
					continue
				}
				log.Printf("prometheus::mesos: %s (port %d) had labels", task.Name, port.Number)
				for _, label := range port.Labels.Labels {
					if label.Key == "DCOS_METRICS_FORMAT" && *label.Value == "prometheus" {
						log.Printf("prometheus::mesos: %s (port %d) was instrumented with prometheus metrics", task.Name, port.Number)
						// TODO error handling

						taskURL := url.URL{
							Scheme: maUrl.Scheme,
							Host:   net.JoinHostPort(maHost, strconv.Itoa(int(port.Number))),
							Path:   "/metrics",
						}
						log.Printf("prometheus::mesos: task url: %q", taskURL)
						uaa := URLAndAddress{URL: &taskURL, OriginalURL: &taskURL}
						result = append(result, uaa)
					}
				}
			}
		}
	}
	return result, nil
}
