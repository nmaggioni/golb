package golb

import (
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
)

type sessionDetails struct {
	Upstream upstream
	LastUsed time.Time
}

var sessions = make(map[string]sessionDetails)
var sessionsLock sync.Mutex

func (s sessionDetails) isStale() bool {
	return time.Now().After(s.LastUsed.Add(time.Duration(config.Stickiness) * time.Second))
}

func staleSessionsCleaner() {
	for {
		for ip, session := range sessions {
			if session.isStale() {
				if config.Verbose {
					logrus.WithFields(logrus.Fields{
						"remote-address": ip,
						"upstream-name":  session.Upstream.Name,
					}).Info("Desticking stale client")
				}
				delete(sessions, ip)
			}
		}
		time.Sleep(1 * time.Second)
	}
}

func updateSession(IP string, session sessionDetails) {
	sessionsLock.Lock()
	defer sessionsLock.Unlock()
	sessions[IP] = session
}

func getSession(IP string) (sessionDetails, bool) {
	sessionsLock.Lock()
	defer sessionsLock.Unlock()
	session, isCached := sessions[IP]
	return session, isCached
}
