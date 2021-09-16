package workers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

type HeartbeatInfo struct {
	Hostname    string   `json:"hostname"`
	StartedAt   int64    `json:"started_at"`
	Pid         int      `json:"pid"`
	Tag         string   `json:"tag"`
	Concurrency int      `json:"concurrency"`
	Queues      []string `json:"queues"`
	Labels      []string `json:"labels"`
	Identity    string   `json:"identity"`
}

type Heartbeat struct {
	Identity string

	Beat  time.Time
	Quiet bool
	Busy  int
	RttUS int
	RSS   int64
	Info  string
}

func (s *apiServer) StartHeartbeat() {
	heartbeatTicker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-heartbeatTicker.C:
			for _, m := range s.managers {
				log.Println("sending heartbeat")
				m.SendHeartbeat()
			}
		}
	}
}

func GenerateProcessNonce() (string, error) {
	bytes := make([]byte, 12)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func BuildHeartbeat(m *Manager) *Heartbeat {
	queues := []string{}
	concurrency := 0
	busy := 0
	for _, w := range m.workers {
		queues = append(queues, w.queue)
		concurrency += w.concurrency // add up all concurrency here because it can be specified on a per-worker basis.
	}

	hostname, _ := os.Hostname()
	pid := os.Getpid()

	if m.opts.ManagerDisplayName != "" {
		hostname = hostname + ":" + m.opts.ManagerDisplayName
	}

	// identity := m.opts.Namespace

	tag := "default"

	if m.opts.Namespace != "" {
		tag = strings.ReplaceAll(m.opts.Namespace, ":", "")
	}

	identity := fmt.Sprintf("%s:%s:%s", hostname, string(pid), m.processNonce)

	h1 := &HeartbeatInfo{
		Hostname:    hostname,
		StartedAt:   m.startedAt.UTC().Unix(),
		Pid:         pid,
		Tag:         tag,
		Concurrency: concurrency,
		Queues:      queues,
		Labels:      []string{},
		Identity:    identity,
	}
	h1m, _ := json.Marshal(h1)

	// inProgress := m.inProgressMessages()
	// ns := m.opts.Namespace

	// for queue, msgs := range inProgress {
	// 	var jobs []JobStatus
	// 	for _, m := range msgs {
	// 		jobs = append(jobs, JobStatus{
	// 			Message:   m,
	// 			StartedAt: m.startedAt,
	// 		})
	// 	}
	// 	stats.Jobs[ns+queue] = jobs
	// 	q = append(q, queue)
	// }

	h := &Heartbeat{
		Identity: identity,
		Beat:     time.Now(),
		Quiet:    false,
		Busy:     busy,
		RSS:      0, // rss is not currently supported
		Info:     string(h1m),
	}

	return h
}
