func (h *Handler) processProxyMode(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	res, err := http.Get(h.target + r.URL.RequestURI())
	if err != nil {
		httpruntime.SetError(r.Context(), r, w, errors.Wrap(err, "couldn't load file"))
		return
	}
	response, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		httpruntime.SetError(r.Context(), r, w, errors.Wrap(err, "couldn't read file content"))
		return
	}
	_, err = w.Write(response)
	if err != nil {
		httpruntime.SetError(r.Context(), r, w, errors.Wrap(err, "couldn't write response"))
		return
	}
}
