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

	maUrl, err := url.Parse(mesosAgentUrl)
	if err != nil {
		return result, err
	}
	maHost, _, err := net.SplitHostPort(maUrl.Host)
	if err != nil {
		return result, err
	}

	// TODO: timeout
	uri := mesosAgentUrl + "/api/v1"
	cli := httpagent.NewSender(httpcli.New(httpcli.Endpoint(uri)).Send)
	ctx := context.Background()

	resp, err := cli.Send(ctx, calls.NonStreaming(calls.GetTasks()))
	if err != nil {
		return result, err
	}

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
		for _, task := range gt.LaunchedTasks {
			if task.Discovery == nil {
				continue
			}
			if task.Discovery.Ports == nil {
				continue
			}
			for _, port := range task.Discovery.Ports.Ports {
				if port.Labels == nil {
					continue
				}
				for _, label := range port.Labels.Labels {
					if label.Key == "DCOS_METRICS_FORMAT" && *label.Value == "prometheus" {
						log.Printf("prometheus: instrumenting mesos task %s (%s:%d) with prometheus metrics", task.Name, maHost, port.Number)

						taskURL := url.URL{
							Scheme: maUrl.Scheme,
							Host:   net.JoinHostPort(maHost, strconv.Itoa(int(port.Number))),
							Path:   "/metrics",
						}
						// TODO address should be the mesos-dns address eg sleep.marathon.mesos
						uaa := URLAndAddress{URL: &taskURL, OriginalURL: &taskURL}
						result = append(result, uaa)
					}
				}
			}
		}
	}
	return result, nil
}
