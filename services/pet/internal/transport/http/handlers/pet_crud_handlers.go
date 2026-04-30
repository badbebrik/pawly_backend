package handlers

import (
	"net/http"
	"pet/internal/application/usecase"
	"pet/internal/transport/http/dto"
)

func (h *Handlers) CreatePet(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	var req dto.CreatePetRequest
	if !decodeRequest(w, r, &req) {
		return
	}

	speciesID, err := parseOptionalUUID(req.SpeciesID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid species_id")
		return
	}
	breedID, err := parseOptionalUUID(req.BreedID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid breed_id")
		return
	}
	patternID, err := parseOptionalUUID(req.PatternID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid pattern_id")
		return
	}
	colors, err := parseColorRequests(req.Colors)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid colors")
		return
	}

	birthDate, err := parseOptionalDate(req.BirthDate, "2006-01-02", "birth_date")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", err.Error())
		return
	}
	microchipInstalledAt, err := parseOptionalDate(req.MicrochipInstalledAt, "2006-01-02", "microchip_installed_at")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", err.Error())
		return
	}

	pet, err := h.useCases.CreatePet(r.Context(), usecase.CreatePetParams{
		UserID:               userID,
		Name:                 req.Name,
		SpeciesID:            speciesID,
		CustomSpeciesName:    req.CustomSpeciesName,
		Sex:                  req.Sex,
		BirthDate:            birthDate,
		BreedID:              breedID,
		CustomBreedName:      req.CustomBreedName,
		Colors:               colors,
		PatternID:            patternID,
		CustomPatternName:    req.CustomPatternName,
		IsNeutered:           req.IsNeutered,
		IsOutdoor:            req.IsOutdoor,
		MicrochipID:          req.MicrochipID,
		MicrochipInstalledAt: microchipInstalledAt,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, dto.PetEnvelopeResponse{
		Pet: petToResponse(pet, h.getPetPhotoDownloadURL(r, pet)),
	})
}

func (h *Handlers) ListPets(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	includeArchived := r.URL.Query().Get("include_archived") == "true"
	offset := parseIntOrDefault(r.URL.Query().Get("offset"), 0)
	limit := parseIntOrDefault(r.URL.Query().Get("limit"), 50)

	items, total, err := h.useCases.ListPets(r.Context(), usecase.ListPetsParams{
		UserID:          userID,
		IncludeArchived: includeArchived,
		Offset:          offset,
		Limit:           limit,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, listPetsToResponse(items, total, offset, limit))
}

func (h *Handlers) GetPet(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	petID, ok := parseRouteUUID(w, r, "pet_id", "pet_id")
	if !ok {
		return
	}

	pet, err := h.useCases.GetPet(r.Context(), userID, petID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.PetEnvelopeResponse{
		Pet: petToResponse(pet, h.getPetPhotoDownloadURL(r, pet)),
	})
}

func (h *Handlers) UpdatePet(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	petID, ok := parseRouteUUID(w, r, "pet_id", "pet_id")
	if !ok {
		return
	}

	var req dto.UpdatePetRequest
	if !decodeRequest(w, r, &req) {
		return
	}

	speciesID, err := parseOptionalUUID(req.SpeciesID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid species_id")
		return
	}
	breedID, err := parseOptionalUUID(req.BreedID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid breed_id")
		return
	}
	patternID, err := parseOptionalUUID(req.PatternID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid pattern_id")
		return
	}
	colors, err := parseColorRequests(req.Colors)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid colors")
		return
	}

	birthDate, err := parseOptionalDate(req.BirthDate, "2006-01-02", "birth_date")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", err.Error())
		return
	}
	microchipInstalledAt, err := parseOptionalDate(req.MicrochipInstalledAt, "2006-01-02", "microchip_installed_at")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", err.Error())
		return
	}

	pet, err := h.useCases.UpdatePet(r.Context(), usecase.UpdatePetParams{
		UserID:               userID,
		PetID:                petID,
		RowVersion:           req.RowVersion,
		Name:                 req.Name,
		SpeciesID:            speciesID,
		CustomSpeciesName:    req.CustomSpeciesName,
		Sex:                  req.Sex,
		BirthDate:            birthDate,
		BreedID:              breedID,
		CustomBreedName:      req.CustomBreedName,
		Colors:               colors,
		PatternID:            patternID,
		CustomPatternName:    req.CustomPatternName,
		IsNeutered:           req.IsNeutered,
		IsOutdoor:            req.IsOutdoor,
		MicrochipID:          req.MicrochipID,
		MicrochipInstalledAt: microchipInstalledAt,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.PetEnvelopeResponse{
		Pet: petToResponse(pet, h.getPetPhotoDownloadURL(r, pet)),
	})
}
