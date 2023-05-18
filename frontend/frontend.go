package frontend

import (
	"html/template"
	"log"
	"mime"
	"net/http"

	"github.com/geniusdex/broawp/accrace"
	"github.com/gorilla/sessions"
)

type frontend struct {
	state        *accrace.State
	templateData *templateData

	websockets *webSocketBroadcaster
}

type templateData struct {
	basePath string
}

func (f *frontend) addTemplateFunctions(t *template.Template, templateData *templateData) *template.Template {
	t.Funcs(template.FuncMap{
		// Environment
		"basePath": func() string {
			return templateData.basePath
		},
	})
	return t
}

func (f *frontend) initializeTemplates() (*template.Template, error) {
	return f.addTemplateFunctions(template.New("templates"), f.templateData).ParseGlob("templates/*.html")
}

func (f *frontend) getTemplates() (*template.Template, error) {
	return f.initializeTemplates()
}

func basePath(r *http.Request) string {
	return r.Header.Get("X-Forwarded-Prefix")
}

func (f *frontend) executeTemplate(w http.ResponseWriter, r *http.Request, name string, data interface{}) {
	sessions.Save(r, w)

	f.templateData.basePath = basePath(r)

	t, err := f.getTemplates()
	if err != nil {
		log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	err = t.ExecuteTemplate(w, name, data)
	if err != nil {
		log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func addMimeTypes() error {
	if err := mime.AddExtensionType(".js", "text/javascript"); err != nil {
		return err
	}

	if err := mime.AddExtensionType(".mjs", "text/javascript"); err != nil {
		return err
	}

	return nil
}

func Run(state *accrace.State) error {
	addMimeTypes()

	f := &frontend{
		state:        state,
		templateData: &templateData{},
		websockets:   newWebSocketBroadcaster(),
	}

	go f.sendWebSocketUpdates()

	http.HandleFunc("/", f.indexHandler)
	http.HandleFunc("/broadcast", f.broadcastHandler)
	http.HandleFunc("/disconnect", f.disconnectHandler)
	http.HandleFunc("/driver", f.driverHandler)
	http.HandleFunc("/overlay", f.overlayHandler)
	http.HandleFunc("/ws", f.webSocketHandler)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	return http.ListenAndServe("localhost:9123", nil)
}
