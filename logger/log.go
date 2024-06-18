package logger

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"go.uber.org/fx/fxevent"
)

type Logger struct {
	*logrus.Logger
}

const (
	bodyFieldName = "dump"
	EnvLevel      = "LOG_LEVEL"
)

func New() *Logger {
	inst := logrus.New()

	inst.SetFormatter(&logrus.TextFormatter{})

	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	multiWriter := io.MultiWriter(file, os.Stdout)
	log.SetOutput(multiWriter)

	inst.SetLevel(logrus.DebugLevel)
	inst.Warningf("Log EnvLevel: %s", logrus.DebugLevel.String())

	return &Logger{
		Logger: inst,
	}
}

func (inst *Logger) HttpReq(req *http.Request) {
	dump, err := httputil.DumpRequest(req, true)
	if err != nil {
		inst.Error(fmt.Errorf("can't make dump of request: %w", err))
	}

	inst.WithField(bodyFieldName, string(dump)).Infof("Request: %s %s", req.Method, req.Host+req.URL.Path)
}

func (inst *Logger) HttpResp(resp *http.Response) {
	dump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		inst.Error(fmt.Errorf("can't make dump of response: %w", err))
	}

	inst.WithField(bodyFieldName, string(dump)).Infof("Response %d", resp.StatusCode)
}

func NewFxEventLogger(log *Logger) fxevent.Logger {
	return &FxEvent{
		log: log,
	}
}

type FxEvent struct {
	log *Logger
}

func (inst *FxEvent) LogEvent(event fxevent.Event) {
	switch e := event.(type) {
	case *fxevent.OnStartExecuting:
		inst.log.Infof("OnStart hook executing %s", e.CallerName)
	case *fxevent.OnStartExecuted:
		if e.Err != nil {
			inst.log.Errorf("OnStart hook failed %s err: %s", e.CallerName, e.Err)
		} else {
			inst.log.Infof("OnStart hook executed %s, runtime %s", e.CallerName, e.Runtime.String())
		}
	case *fxevent.OnStopExecuting:
		inst.log.Infof("OnStop hook executing %s", e.CallerName)
	case *fxevent.OnStopExecuted:
		if e.Err != nil {
			inst.log.Errorf("OnStop hook failed %s err: %s", e.CallerName, e.Err)
		} else {
			inst.log.Infof("OnStop hook executed %s, runtime %s", e.CallerName, e.Runtime.String())
		}
	case *fxevent.Supplied:
		if e.Err != nil {
			inst.log.Errorf("Supply %s error: %s", e.TypeName, e.Err)
		} else {
			inst.log.Infof("Supply  %s", e.TypeName)
		}
	case *fxevent.Provided:
		for _, rtype := range e.OutputTypeNames {
			inst.log.Infof("Provide %s", rtype)
		}
		if e.Err != nil {
			inst.log.Errorf("Error encountered while applying options: %s", e.Err)
		}
	case *fxevent.Invoking:
		// Do not log stack as it will make logs hard to read.
		inst.log.Infof("Invoking %s", e.FunctionName)
	case *fxevent.Invoked:
		if e.Err != nil {
			inst.log.WithField("stack", e.Trace).Errorf("Invoke failed %s err: %s", e.FunctionName, e.Err)
		}
	case *fxevent.Stopping:
		inst.log.Warnf("Received signal: %s", strings.ToUpper(e.Signal.String()))
	case *fxevent.Stopped:
		if e.Err != nil {
			inst.log.Errorf("Stop failed, err: %s", e.Err.Error())
		}
	case *fxevent.RollingBack:
		inst.log.Errorf("Start failed, rolling back: err %s", e.StartErr)
	case *fxevent.RolledBack:
		if e.Err != nil {
			inst.log.Errorf("Rollback failed, err: %s", e.Err)
		}
	case *fxevent.Started:
		if e.Err != nil {
			inst.log.Errorf("Start failed: %s", e.Err)
		} else {
			inst.log.Info("Started")
		}
	case *fxevent.LoggerInitialized:
		if e.Err != nil {
			inst.log.Errorf("Custom logger initialization failed, err: %s", e.Err)
		} else {
			inst.log.Infof("Initialized custom fxevent.Logger %s", e.ConstructorName)
		}
	}
}
