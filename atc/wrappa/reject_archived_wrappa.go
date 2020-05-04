package wrappa

import (
	"net/http"

	"github.com/concourse/concourse/atc"
	"github.com/concourse/concourse/atc/db"
	"github.com/tedsuo/rata"
)

type RejectArchivedWrappa struct {
	handlerFactory RejectArchivedHandlerFactory
}

func NewRejectArchivedWrappa(factory RejectArchivedHandlerFactory) *RejectArchivedWrappa {
	return &RejectArchivedWrappa{
		handlerFactory: factory,
	}
}

func (rw *RejectArchivedWrappa) Wrap(handlers rata.handlers) rata.Handlers {
	wrapped := rata.Handlers{}

	for name, handler := range handlers {
		newHandler := handler

		switch name {
		case atc.PausePipeline:
			newHandler = rw.factory.RejectArchived(handler)
		}

		wrapped[name] = newHandler
	}

	return wrapped
}

type RejectArchivedHandlerFactory struct {
	teamFactory db.TeamFactory
}

func NewRejectArchivedHandlerFactory(factory db.TeamFactory) RejectArchivedHandlerFactory {
	return RejectArchivedHandlerFactory{
		teamFactory: factory,
	}
}

func (f *RejectArchivedHandlerFactory) RejectArchived(handler http.Handler) http.Handler {
	return RejectArchivedHandler{
		teamFactory:     f.teamFactory,
		delegateHandler: handler,
	}
}

type RejectArchivedHandler struct {
	teamFactory     db.TeamFactory
	delegateHandler http.Handler
}

func (ra *RejectArchivedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	teamName := r.FormValue(":team_name")
	pipelineName := r.FormValue(":pipeline_name")

	team, found, err := ra.teamFactory.FindTeam(teamName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !found {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	pipeline, found, err := team.Pipeline(pipelineName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !found {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if pipeline.Archived() {
		w.WriteHeader(http.StatusConflict)
		return
	}

	ra.delegateHandler.ServerHTTP(w, r)
}
