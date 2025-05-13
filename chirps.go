package main

func (cfg *apiConfig) setChirp(res http.ResponseWriter, req *http.Request) {
	requestParams := database.CreateChirpParams{UpdatedAt: time.Now(), CreatedAt: time.Now(), ID: uuid.New()}

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&requestParams)
	if err != nil {
		fmt.Println(err)
	}
	chrp, err := validateChirp(requestParams)
	if err != nil {
		fmt.Println(err)
	}
	chrpDB, err := cfg.db.CreateChirp(req.Context(), chrp)
	if err != nil {
		fmt.Println(err)
	}
	respondWithJSON(res, 201, chrpDB)
}
