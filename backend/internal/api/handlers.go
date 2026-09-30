package api

import (
	"net/http"

	"opstasks/internal/store"
)

func (s *Server) health(r *http.Request) (any, error) {
	if err := s.store.Ping(r.Context()); err != nil {
		return nil, err
	}
	return map[string]string{"status": "ok", "today": s.clock.Today()}, nil
}

func (s *Server) dashboard(r *http.Request) (any, error) {
	d, err := s.store.Dashboard(r.Context(), s.clock.Today())
	if err != nil {
		return nil, err
	}
	d.Timezone = s.clock.Loc.String()
	return d, nil
}

// history returns the tasks completed on ?day= (default today) and daily counts for
// ?month= (default the day's month).
func (s *Server) history(r *http.Request) (any, error) {
	q := r.URL.Query()
	day := q.Get("day")
	if day == "" {
		day = s.clock.Today()
	}
	return s.store.History(r.Context(), day, q.Get("month"))
}

// categories

func (s *Server) listCategories(r *http.Request) (any, error) {
	return s.store.ListCategories(r.Context())
}

func (s *Server) createCategory(r *http.Request) (any, error) {
	var in store.CategoryInput
	if err := decode(r, &in); err != nil {
		return nil, err
	}
	return s.store.CreateCategory(r.Context(), in)
}

func (s *Server) updateCategory(r *http.Request) (any, error) {
	id, err := pathID(r)
	if err != nil {
		return nil, err
	}
	var in store.CategoryInput
	if err := decode(r, &in); err != nil {
		return nil, err
	}
	return s.store.UpdateCategory(r.Context(), id, in)
}

func (s *Server) deleteCategory(r *http.Request) (any, error) {
	id, err := pathID(r)
	if err != nil {
		return nil, err
	}
	return nil, s.store.DeleteCategory(r.Context(), id)
}

// tasks

func (s *Server) createTask(r *http.Request) (any, error) {
	var in store.TaskInput
	if err := decode(r, &in); err != nil {
		return nil, err
	}
	return s.store.CreateTask(r.Context(), in, s.clock.Today())
}

func (s *Server) getTask(r *http.Request) (any, error) {
	id, err := pathID(r)
	if err != nil {
		return nil, err
	}
	return s.store.GetTask(r.Context(), id)
}

func (s *Server) updateTask(r *http.Request) (any, error) {
	id, err := pathID(r)
	if err != nil {
		return nil, err
	}
	var in store.TaskInput
	if err := decode(r, &in); err != nil {
		return nil, err
	}
	return s.store.UpdateTask(r.Context(), id, in)
}

func (s *Server) deleteTask(r *http.Request) (any, error) {
	id, err := pathID(r)
	if err != nil {
		return nil, err
	}
	return nil, s.store.DeleteTask(r.Context(), id)
}

func (s *Server) completeTask(r *http.Request) (any, error) {
	id, err := pathID(r)
	if err != nil {
		return nil, err
	}
	return s.store.CompleteTask(r.Context(), id, s.clock.Today())
}

func (s *Server) reopenTask(r *http.Request) (any, error) {
	id, err := pathID(r)
	if err != nil {
		return nil, err
	}
	return s.store.ReopenTask(r.Context(), id)
}

// subtasks

func (s *Server) createSubtask(r *http.Request) (any, error) {
	id, err := pathID(r)
	if err != nil {
		return nil, err
	}
	var in store.SubtaskInput
	if err := decode(r, &in); err != nil {
		return nil, err
	}
	return s.store.CreateSubtask(r.Context(), id, in)
}

func (s *Server) reorderSubtasks(r *http.Request) (any, error) {
	id, err := pathID(r)
	if err != nil {
		return nil, err
	}
	var in struct {
		IDs []int64 `json:"ids"`
	}
	if err := decode(r, &in); err != nil {
		return nil, err
	}
	return s.store.ReorderSubtasks(r.Context(), id, in.IDs)
}

func (s *Server) updateSubtask(r *http.Request) (any, error) {
	id, err := pathID(r)
	if err != nil {
		return nil, err
	}
	var in store.SubtaskInput
	if err := decode(r, &in); err != nil {
		return nil, err
	}
	return s.store.UpdateSubtask(r.Context(), id, in)
}

func (s *Server) deleteSubtask(r *http.Request) (any, error) {
	id, err := pathID(r)
	if err != nil {
		return nil, err
	}
	return nil, s.store.DeleteSubtask(r.Context(), id)
}

// projects

func (s *Server) listProjects(r *http.Request) (any, error) {
	q := r.URL.Query()
	f := store.ProjectsActive
	switch {
	case q.Get("closed") == "1":
		f = store.ProjectsClosed
	case q.Get("all") == "1":
		f = store.ProjectsUnarchived
	}
	return s.store.ListProjects(r.Context(), f)
}

func (s *Server) createProject(r *http.Request) (any, error) {
	var in store.ProjectInput
	if err := decode(r, &in); err != nil {
		return nil, err
	}
	return s.store.CreateProject(r.Context(), in)
}

func (s *Server) getProject(r *http.Request) (any, error) {
	id, err := pathID(r)
	if err != nil {
		return nil, err
	}
	return s.store.GetProject(r.Context(), id)
}

func (s *Server) updateProject(r *http.Request) (any, error) {
	id, err := pathID(r)
	if err != nil {
		return nil, err
	}
	var in store.ProjectInput
	if err := decode(r, &in); err != nil {
		return nil, err
	}
	return s.store.UpdateProject(r.Context(), id, in)
}

func (s *Server) deleteProject(r *http.Request) (any, error) {
	id, err := pathID(r)
	if err != nil {
		return nil, err
	}
	return nil, s.store.DeleteProject(r.Context(), id)
}
