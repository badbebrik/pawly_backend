package usecase

import (
	"context"
	"health/internal/application/ports"
	"health/internal/domain/model"
	"sort"
	"time"

	"github.com/google/uuid"
)

const maxCalendarRangeDays = 93

type Overview struct {
	acl        ports.HealthAccessChecker
	petLister  ports.HealthPetLister
	scheduled  ports.ScheduledRepository
	calendar   ports.CalendarRepository
	dictionary ports.HealthDictionaryRepository
}

func NewOverview(acl ports.HealthAccessChecker, petLister ports.HealthPetLister, scheduled ports.ScheduledRepository, calendar ports.CalendarRepository, dictionary ports.HealthDictionaryRepository) *Overview {
	return &Overview{acl: acl, petLister: petLister, scheduled: scheduled, calendar: calendar, dictionary: dictionary}
}

func (u *Overview) GetHealthBootstrap(ctx context.Context, userID, petID uuid.UUID) (*model.HealthBootstrap, error) {
	if userID == uuid.Nil || petID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	readAllowed, err := u.acl.Check(ctx, petID, userID, ActionHealthRead)
	if err != nil {
		return nil, err
	}
	if !readAllowed {
		return nil, ErrForbidden
	}
	writeAllowed, err := u.acl.Check(ctx, petID, userID, ActionHealthWrite)
	if err != nil {
		return nil, err
	}
	logAllowed, err := u.acl.Check(ctx, petID, userID, ActionLogRead)
	if err != nil {
		return nil, err
	}
	dictionaryItems, err := u.dictionary.ListHealthDictionaryItems(ctx, ports.ListHealthDictionaryItemsInput{
		PetID: petID,
		Kinds: []string{
			ports.HealthDictionaryKindProcedureType,
			ports.HealthDictionaryKindMedicalRecordType,
			ports.HealthDictionaryKindVaccinationTarget,
		},
	})
	if err != nil {
		return nil, mapRepoErr(err)
	}
	return &model.HealthBootstrap{
		Permissions: model.HealthPermissions{HealthRead: true, HealthWrite: writeAllowed, LogRead: logAllowed},
		Enums: model.HealthEnums{
			VetVisitStatuses:       []string{"PLANNED", "COMPLETED"},
			VetVisitTypes:          []string{"CHECKUP", "SYMPTOM", "FOLLOW_UP", "VACCINATION", "PROCEDURE", "OTHER"},
			VaccinationStatuses:    []string{"PLANNED", "COMPLETED"},
			VaccinationTargets:     dictionaryItemsByKind(dictionaryItems, ports.HealthDictionaryKindVaccinationTarget),
			ProcedureStatuses:      []string{"PLANNED", "COMPLETED"},
			ProcedureTypeItems:     dictionaryItemsByKind(dictionaryItems, ports.HealthDictionaryKindProcedureType),
			MedicalRecordTypeItems: dictionaryItemsByKind(dictionaryItems, ports.HealthDictionaryKindMedicalRecordType),
			MedicalRecordStatuses:  []string{"ACTIVE", "RESOLVED"},
		},
	}, nil
}

func (u *Overview) GetHealthDay(ctx context.Context, userID, petID uuid.UUID, day time.Time) ([]model.CalendarDayItem, error) {
	if userID == uuid.Nil || petID == uuid.Nil || day.IsZero() {
		return nil, ErrInvalidInput
	}
	access, err := u.getCalendarPetAccess(ctx, petID, userID)
	if err != nil {
		return nil, err
	}
	sourceTypes := scheduledReadSourceTypes(access)
	if len(sourceTypes) == 0 {
		return nil, ErrForbidden
	}
	dayStart := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC)
	dayEnd := dayStart.Add(24*time.Hour - time.Nanosecond)
	occurrences := []model.ScheduledItemOccurrenceListItem{}
	if len(sourceTypes) > 0 {
		occurrences, err = u.scheduled.ListCalendarDayScheduledOccurrences(ctx, petID, dayStart, dayEnd)
		if err != nil {
			return nil, mapRepoErr(err)
		}
		occurrences = filterScheduledOccurrencesByAccess(occurrences, map[uuid.UUID]ports.PetAccess{petID: access})
	}
	facts := []model.CalendarDayItem{}
	if access.HealthRead {
		facts, err = u.calendar.ListCalendarDayMedicalFacts(ctx, []uuid.UUID{petID}, dayStart, dayEnd)
		if err != nil {
			return nil, mapRepoErr(err)
		}
	}
	items := make([]model.CalendarDayItem, 0, len(occurrences)+len(facts))
	for i := range occurrences {
		items = append(items, scheduledOccurrenceToCalendarDayItem(occurrences[i]))
	}
	items = append(items, facts...)
	sort.SliceStable(items, func(i, j int) bool {
		if !items[i].ScheduledFor.Equal(items[j].ScheduledFor) {
			return items[i].ScheduledFor.Before(items[j].ScheduledFor)
		}
		if items[i].PetID != items[j].PetID {
			return items[i].PetID.String() < items[j].PetID.String()
		}
		if items[i].ItemType != items[j].ItemType {
			return items[i].ItemType < items[j].ItemType
		}
		return items[i].EntityID.String() < items[j].EntityID.String()
	})
	return items, nil
}

func (u *Overview) GetGlobalHealthDay(ctx context.Context, userID uuid.UUID, day time.Time) ([]model.CalendarDayItem, error) {
	if userID == uuid.Nil || day.IsZero() {
		return nil, ErrInvalidInput
	}
	accesses, err := u.petLister.ListPetAccessForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(accesses) == 0 {
		return []model.CalendarDayItem{}, nil
	}
	accessByPet := make(map[uuid.UUID]ports.PetAccess, len(accesses))
	scheduledPetIDs := make([]uuid.UUID, 0, len(accesses))
	healthPetIDs := make([]uuid.UUID, 0, len(accesses))
	for i := range accesses {
		access := accesses[i]
		if len(scheduledReadSourceTypes(access)) > 0 {
			accessByPet[access.PetID] = access
			scheduledPetIDs = append(scheduledPetIDs, access.PetID)
		}
		if access.HealthRead {
			healthPetIDs = append(healthPetIDs, access.PetID)
		}
	}
	dayStart := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC)
	dayEnd := dayStart.Add(24*time.Hour - time.Nanosecond)
	occurrences := []model.ScheduledItemOccurrenceListItem{}
	if len(scheduledPetIDs) > 0 {
		occurrences, err = u.scheduled.ListCalendarDayScheduledOccurrencesForPets(ctx, scheduledPetIDs, dayStart, dayEnd)
		if err != nil {
			return nil, mapRepoErr(err)
		}
		occurrences = filterScheduledOccurrencesByAccess(occurrences, accessByPet)
	}
	facts := []model.CalendarDayItem{}
	if len(healthPetIDs) > 0 {
		facts, err = u.calendar.ListCalendarDayMedicalFacts(ctx, healthPetIDs, dayStart, dayEnd)
		if err != nil {
			return nil, mapRepoErr(err)
		}
	}
	items := make([]model.CalendarDayItem, 0, len(occurrences)+len(facts))
	for i := range occurrences {
		items = append(items, scheduledOccurrenceToCalendarDayItem(occurrences[i]))
	}
	items = append(items, facts...)
	sort.SliceStable(items, func(i, j int) bool {
		if !items[i].ScheduledFor.Equal(items[j].ScheduledFor) {
			return items[i].ScheduledFor.Before(items[j].ScheduledFor)
		}
		if items[i].PetID != items[j].PetID {
			return items[i].PetID.String() < items[j].PetID.String()
		}
		if items[i].ItemType != items[j].ItemType {
			return items[i].ItemType < items[j].ItemType
		}
		return items[i].EntityID.String() < items[j].EntityID.String()
	})
	return items, nil
}

func (u *Overview) GetGlobalHealthCalendar(ctx context.Context, userID uuid.UUID, dateFrom, dateTo time.Time) ([]model.CalendarDateMarker, error) {
	if userID == uuid.Nil || dateFrom.IsZero() || dateTo.IsZero() {
		return nil, ErrInvalidInput
	}
	rangeStart := calendarDateStart(dateFrom)
	rangeEndDate := calendarDateStart(dateTo)
	if rangeEndDate.Before(rangeStart) {
		return nil, ErrInvalidInput
	}
	if int(rangeEndDate.Sub(rangeStart).Hours()/24)+1 > maxCalendarRangeDays {
		return nil, ErrInvalidInput
	}

	accesses, err := u.petLister.ListPetAccessForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(accesses) == 0 {
		return []model.CalendarDateMarker{}, nil
	}
	accessByPet := make(map[uuid.UUID]ports.PetAccess, len(accesses))
	scheduledPetIDs := make([]uuid.UUID, 0, len(accesses))
	healthPetIDs := make([]uuid.UUID, 0, len(accesses))
	for i := range accesses {
		access := accesses[i]
		if len(scheduledReadSourceTypes(access)) > 0 {
			accessByPet[access.PetID] = access
			scheduledPetIDs = append(scheduledPetIDs, access.PetID)
		}
		if access.HealthRead {
			healthPetIDs = append(healthPetIDs, access.PetID)
		}
	}

	rangeEnd := rangeEndDate.Add(24*time.Hour - time.Nanosecond)
	markersByDate := map[string]*model.CalendarDateMarker{}
	addMarker := func(scheduledFor time.Time, status string) {
		date := calendarDateStart(scheduledFor)
		key := date.Format("2006-01-02")
		marker := markersByDate[key]
		if marker == nil {
			marker = &model.CalendarDateMarker{Date: date}
			markersByDate[key] = marker
		}
		if status == "COMPLETED" {
			marker.CompletedCount++
		} else {
			marker.PlannedCount++
		}
		marker.TotalCount++
	}

	if len(scheduledPetIDs) > 0 {
		occurrences, err := u.scheduled.ListCalendarDayScheduledOccurrencesForPets(ctx, scheduledPetIDs, rangeStart, rangeEnd)
		if err != nil {
			return nil, mapRepoErr(err)
		}
		occurrences = filterScheduledOccurrencesByAccess(occurrences, accessByPet)
		for i := range occurrences {
			addMarker(occurrences[i].ScheduledFor, "PLANNED")
		}
	}
	if len(healthPetIDs) > 0 {
		facts, err := u.calendar.ListCalendarDayMedicalFacts(ctx, healthPetIDs, rangeStart, rangeEnd)
		if err != nil {
			return nil, mapRepoErr(err)
		}
		for i := range facts {
			addMarker(facts[i].ScheduledFor, facts[i].Status)
		}
	}

	markers := make([]model.CalendarDateMarker, 0, len(markersByDate))
	for _, marker := range markersByDate {
		markers = append(markers, *marker)
	}
	sort.SliceStable(markers, func(i, j int) bool {
		return markers[i].Date.Before(markers[j].Date)
	})
	return markers, nil
}

func (u *Overview) getCalendarPetAccess(ctx context.Context, petID, userID uuid.UUID) (ports.PetAccess, error) {
	var access ports.PetAccess
	access.PetID = petID
	checks := []struct {
		action string
		set    func(bool)
	}{
		{ActionPetRead, func(v bool) { access.PetRead = v }},
		{ActionLogRead, func(v bool) { access.LogRead = v }},
		{ActionHealthRead, func(v bool) { access.HealthRead = v }},
	}
	for i := range checks {
		allowed, err := u.acl.Check(ctx, petID, userID, checks[i].action)
		if err != nil {
			return ports.PetAccess{}, err
		}
		checks[i].set(allowed)
	}
	return access, nil
}

func calendarDateStart(value time.Time) time.Time {
	utc := value.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}

func filterScheduledOccurrencesByAccess(items []model.ScheduledItemOccurrenceListItem, accessByPet map[uuid.UUID]ports.PetAccess) []model.ScheduledItemOccurrenceListItem {
	if len(items) == 0 {
		return items
	}
	filtered := items[:0]
	for i := range items {
		access, ok := accessByPet[items[i].PetID]
		if !ok {
			continue
		}
		if scheduledSourceReadable(access, items[i].Rule.SourceType) {
			filtered = append(filtered, items[i])
		}
	}
	return filtered
}

func scheduledSourceReadable(access ports.PetAccess, sourceType string) bool {
	switch sourceType {
	case model.ScheduledItemSourceTypeManual, model.ScheduledItemSourceTypePetEvent:
		return access.PetRead
	case model.ScheduledItemSourceTypeLogType:
		return access.LogRead
	case model.ScheduledItemSourceTypeVetVisit, model.ScheduledItemSourceTypeVaccination, model.ScheduledItemSourceTypeProcedure:
		return access.HealthRead
	default:
		return false
	}
}

func scheduledOccurrenceToCalendarDayItem(item model.ScheduledItemOccurrenceListItem) model.CalendarDayItem {
	itemType := item.Rule.SourceType
	entityID := item.Rule.ID
	if item.Rule.SourceID != nil {
		entityID = *item.Rule.SourceID
	}
	out := model.CalendarDayItem{
		ItemType:              itemType,
		EntityID:              entityID,
		PetID:                 item.PetID,
		Title:                 item.Rule.Title,
		Subtitle:              item.Rule.Note,
		ScheduledFor:          item.ScheduledFor,
		Status:                "PLANNED",
		ScheduledItemID:       &item.ScheduledItemID,
		ScheduledOccurrenceID: &item.ID,
	}
	switch itemType {
	case model.ScheduledItemSourceTypeVetVisit:
		out.VisitID = item.Rule.SourceID
	case model.ScheduledItemSourceTypeVaccination:
		out.VaccinationID = item.Rule.SourceID
	case model.ScheduledItemSourceTypeProcedure:
		out.ProcedureID = item.Rule.SourceID
	}
	return out
}
