package handlers

import (
	domainmodel "health/internal/domain/model"
	"health/internal/transport/http/dto"
	"time"

	"github.com/google/uuid"
)

func healthBootstrapToDTO(item *domainmodel.HealthBootstrap) dto.HealthBootstrapResponse {
	return dto.HealthBootstrapResponse{
		Permissions: dto.HealthBootstrapPermissionsResponse{
			HealthRead:  item.Permissions.HealthRead,
			HealthWrite: item.Permissions.HealthWrite,
			LogRead:     item.Permissions.LogRead,
		},
		Enums: dto.HealthBootstrapEnumsResponse{
			VetVisitStatuses:       item.Enums.VetVisitStatuses,
			VetVisitTypes:          item.Enums.VetVisitTypes,
			VaccinationStatuses:    item.Enums.VaccinationStatuses,
			VaccinationTargets:     healthDictionaryItemsToDTO(item.Enums.VaccinationTargets),
			ProcedureStatuses:      item.Enums.ProcedureStatuses,
			ProcedureTypeItems:     healthDictionaryItemsToDTO(item.Enums.ProcedureTypeItems),
			MedicalRecordTypeItems: healthDictionaryItemsToDTO(item.Enums.MedicalRecordTypeItems),
			MedicalRecordStatuses:  item.Enums.MedicalRecordStatuses,
		},
	}
}

func healthDictionaryItemToDTO(item domainmodel.HealthDictionaryItem) dto.HealthDictionaryItemResponse {
	return dto.HealthDictionaryItemResponse{
		ID:         item.ID.String(),
		Kind:       item.Kind,
		PetID:      uuidStringPtr(item.PetID),
		Code:       item.Code,
		Name:       item.Name,
		IsSystem:   item.IsSystem,
		IsArchived: item.IsArchived,
	}
}

func healthDictionaryItemPtrToDTO(item *domainmodel.HealthDictionaryItem) *dto.HealthDictionaryItemResponse {
	if item == nil {
		return nil
	}
	out := healthDictionaryItemToDTO(*item)
	return &out
}

func healthDictionaryItemsToDTO(items []domainmodel.HealthDictionaryItem) []dto.HealthDictionaryItemResponse {
	out := make([]dto.HealthDictionaryItemResponse, 0, len(items))
	for i := range items {
		out = append(out, healthDictionaryItemToDTO(items[i]))
	}
	return out
}

func logAppListItemToDTO(item domainmodel.LogListItem) dto.LogListItemResponse {
	metricValues := make([]dto.LogMetricValueResponse, 0, len(item.MetricValues))
	for i := range item.MetricValues {
		mv := item.MetricValues[i]
		metricValues = append(metricValues, dto.LogMetricValueResponse{
			MetricID:   mv.MetricID.String(),
			MetricName: mv.MetricName,
			InputKind:  mv.InputKind,
			Unit:       mv.Unit,
			ValueNum:   mv.ValueNum,
		})
	}

	return dto.LogListItemResponse{
		ID:                  item.ID.String(),
		PetID:               item.PetID.String(),
		OccurredAt:          item.OccurredAt.UTC().Format(time.RFC3339),
		LogTypeID:           uuidStringPtr(item.LogTypeID),
		LogTypeName:         item.LogTypeName,
		LogTypeScope:        item.LogTypeScope,
		DescriptionPreview:  item.DescriptionPreview,
		RelatedEntityType:   item.RelatedEntityType,
		RelatedEntityID:     uuidStringPtr(item.RelatedEntityID),
		MetricValuesPreview: metricValues,
		AttachmentsCount:    item.AttachmentsCount,
		HasAttachments:      item.HasAttachments,
		CreatedByUserID:     item.CreatedByUserID.String(),
	}
}

func logAppToDTO(item *domainmodel.Log) dto.LogResponse {
	metricValues := make([]dto.LogMetricValueResponse, 0, len(item.MetricValues))
	for i := range item.MetricValues {
		mv := item.MetricValues[i]
		metricValues = append(metricValues, dto.LogMetricValueResponse{
			MetricID:   mv.MetricID.String(),
			MetricName: mv.MetricName,
			InputKind:  mv.InputKind,
			Unit:       mv.Unit,
			ValueNum:   mv.ValueNum,
		})
	}

	attachments := make([]dto.LogAttachmentResponse, 0, len(item.Attachments))
	for i := range item.Attachments {
		a := item.Attachments[i]
		attachments = append(attachments, dto.LogAttachmentResponse{
			ID:            a.ID.String(),
			FileID:        a.FileID.String(),
			FileName:      a.FileName,
			FileType:      a.FileType,
			DownloadURL:   a.DownloadURL,
			PreviewURL:    a.PreviewURL,
			AddedAt:       a.AddedAt.UTC().Format(time.RFC3339),
			AddedByUserID: a.AddedByUserID.String(),
		})
	}

	return dto.LogResponse{
		ID:                item.ID.String(),
		PetID:             item.PetID.String(),
		OccurredAt:        item.OccurredAt.UTC().Format(time.RFC3339),
		LogTypeID:         uuidStringPtr(item.LogTypeID),
		LogTypeName:       item.LogTypeName,
		LogTypeScope:      item.LogTypeScope,
		Description:       item.Description,
		RelatedEntityType: item.RelatedEntityType,
		RelatedEntityID:   uuidStringPtr(item.RelatedEntityID),
		MetricValues:      metricValues,
		Attachments:       attachments,
		RowVersion:        item.RowVersion,
		CreatedAt:         item.CreatedAt.UTC().Format(time.RFC3339),
		CreatedByUserID:   item.CreatedByUserID.String(),
		UpdatedAt:         item.UpdatedAt.UTC().Format(time.RFC3339),
		UpdatedByUserID:   item.UpdatedByUserID.String(),
		CanEdit:           true,
		CanDelete:         true,
	}
}

func logTypeAppToDTO(item domainmodel.LogType) dto.LogTypeResponse {
	metricRequirements := make([]dto.LogTypeMetricRequirementResponse, 0, len(item.MetricRequirements))
	for i := range item.MetricRequirements {
		req := item.MetricRequirements[i]
		metricRequirements = append(metricRequirements, dto.LogTypeMetricRequirementResponse{
			MetricID:   req.MetricID.String(),
			IsRequired: req.IsRequired,
		})
	}

	return dto.LogTypeResponse{
		ID:                 item.ID.String(),
		Scope:              item.Scope,
		PetID:              uuidStringPtr(item.PetID),
		Code:               item.Code,
		Name:               item.Name,
		MetricRequirements: metricRequirements,
		CreatedAt:          item.CreatedAt.UTC().Format(time.RFC3339),
		CreatedByUserID:    uuidStringPtr(item.CreatedByUserID),
		UpdatedAt:          item.UpdatedAt.UTC().Format(time.RFC3339),
		UpdatedByUserID:    uuidStringPtr(item.UpdatedByUserID),
		RowVersion:         item.RowVersion,
		IsArchived:         item.DeletedAt != nil,
	}
}

func metricAppToDTO(item domainmodel.Metric) dto.MetricResponse {
	return dto.MetricResponse{
		ID:              item.ID.String(),
		Scope:           item.Scope,
		PetID:           uuidStringPtr(item.PetID),
		Code:            item.Code,
		Name:            item.Name,
		InputKind:       item.InputKind,
		Unit:            item.Unit,
		MinValue:        item.MinValue,
		MaxValue:        item.MaxValue,
		CreatedAt:       item.CreatedAt.UTC().Format(time.RFC3339),
		CreatedByUserID: uuidStringPtr(item.CreatedByUserID),
		UpdatedAt:       item.UpdatedAt.UTC().Format(time.RFC3339),
		UpdatedByUserID: uuidStringPtr(item.UpdatedByUserID),
		RowVersion:      item.RowVersion,
		IsArchived:      item.DeletedAt != nil,
		Usage: dto.MetricUsageResponse{
			LogTypesCount: item.Usage.LogTypesCount,
			LogsCount:     item.Usage.LogsCount,
		},
	}
}

func analyticsMetricSummaryAppToDTO(item domainmodel.AnalyticsMetricSummary) dto.AnalyticsMetricSummaryResponse {
	used := make([]dto.AnalyticsUsedLogTypeResponse, 0, len(item.UsedInLogTypes))
	for i := range item.UsedInLogTypes {
		used = append(used, dto.AnalyticsUsedLogTypeResponse{
			LogTypeID:   item.UsedInLogTypes[i].LogTypeID.String(),
			LogTypeName: item.UsedInLogTypes[i].LogTypeName,
		})
	}
	return dto.AnalyticsMetricSummaryResponse{
		MetricID:        item.MetricID.String(),
		MetricName:      item.MetricName,
		MetricScope:     item.MetricScope,
		InputKind:       item.InputKind,
		Unit:            item.Unit,
		PointsCount:     item.PointsCount,
		FirstOccurredAt: item.FirstOccurredAt.UTC().Format(time.RFC3339),
		LastOccurredAt:  item.LastOccurredAt.UTC().Format(time.RFC3339),
		LastValueNum:    item.LastValueNum,
		UsedInLogTypes:  used,
	}
}

func metricSeriesPointAppToDTO(item domainmodel.MetricSeriesPoint) dto.MetricSeriesPointResponse {
	return dto.MetricSeriesPointResponse{
		OccurredAt:  item.OccurredAt.UTC().Format(time.RFC3339),
		ValueNum:    item.ValueNum,
		LogID:       item.LogID.String(),
		LogTypeID:   uuidStringPtr(item.LogTypeID),
		LogTypeName: item.LogTypeName,
	}
}

func metricSeriesSummaryAppToDTO(item *domainmodel.MetricSeriesSummary) *dto.MetricSeriesSummaryResponse {
	if item == nil {
		return nil
	}
	return &dto.MetricSeriesSummaryResponse{
		PointsCount:       item.PointsCount,
		MinValueNum:       item.MinValueNum,
		MaxValueNum:       item.MaxValueNum,
		LastValueNum:      item.LastValueNum,
		AvgValueNum:       item.AvgValueNum,
		DeltaFromFirstNum: item.DeltaFromFirstNum,
	}
}

func vetVisitAppListItemToDTO(item domainmodel.VetVisitListItem) dto.VetVisitListItemResponse {
	return dto.VetVisitListItemResponse{
		ID:               item.ID.String(),
		PetID:            item.PetID.String(),
		Status:           item.Status,
		VisitType:        item.VisitType,
		Title:            item.Title,
		ScheduledAt:      timeStringPtr(item.ScheduledAt),
		CompletedAt:      timeStringPtr(item.CompletedAt),
		ReasonText:       item.ReasonText,
		ResultText:       item.ResultText,
		ClinicName:       item.ClinicName,
		VetName:          item.VetName,
		RelatedLogsCount: item.RelatedLogsCount,
		AttachmentsCount: item.AttachmentsCount,
		RowVersion:       item.RowVersion,
		CreatedAt:        item.CreatedAt.UTC().Format(time.RFC3339),
		CreatedByUserID:  item.CreatedByUserID.String(),
		UpdatedAt:        item.UpdatedAt.UTC().Format(time.RFC3339),
		UpdatedByUserID:  item.UpdatedByUserID.String(),
	}
}

func vetVisitAppToDTO(item *domainmodel.VetVisit) dto.VetVisitResponse {
	relatedLogs := make([]dto.RelatedLogResponse, 0, len(item.RelatedLogs))
	for i := range item.RelatedLogs {
		relatedLogs = append(relatedLogs, relatedLogAppToDTO(item.RelatedLogs[i]))
	}
	attachments := make([]dto.HealthAttachmentResponse, 0, len(item.Attachments))
	for i := range item.Attachments {
		attachments = append(attachments, healthAttachmentAppToDTO(item.Attachments[i]))
	}
	return dto.VetVisitResponse{
		ID:            item.ID.String(),
		PetID:         item.PetID.String(),
		Status:        item.Status,
		VisitType:     item.VisitType,
		Title:         item.Title,
		ScheduledAt:   timeStringPtr(item.ScheduledAt),
		CompletedAt:   timeStringPtr(item.CompletedAt),
		ReasonText:    item.ReasonText,
		ResultText:    item.ResultText,
		ClinicName:    item.ClinicName,
		VetName:       item.VetName,
		RelatedLogs:   relatedLogs,
		Attachments:   attachments,
		RowVersion:    item.RowVersion,
		CreatedAt:     item.CreatedAt.UTC().Format(time.RFC3339),
		CreatedByUser: item.CreatedByUserID.String(),
		UpdatedAt:     item.UpdatedAt.UTC().Format(time.RFC3339),
		UpdatedByUser: item.UpdatedByUserID.String(),
		CanEdit:       true,
		CanDelete:     true,
	}
}

func healthAttachmentAppToDTO(item domainmodel.HealthAttachment) dto.HealthAttachmentResponse {
	return dto.HealthAttachmentResponse{
		ID:            item.ID.String(),
		FileID:        item.FileID.String(),
		FileName:      item.FileName,
		FileType:      item.FileType,
		DownloadURL:   item.DownloadURL,
		PreviewURL:    item.PreviewURL,
		AddedByUserID: item.AddedByUserID.String(),
		AddedAt:       item.AddedAt.UTC().Format(time.RFC3339),
	}
}

func relatedLogAppToDTO(item domainmodel.RelatedLog) dto.RelatedLogResponse {
	return dto.RelatedLogResponse{
		ID:                 item.ID.String(),
		OccurredAt:         item.OccurredAt.UTC().Format(time.RFC3339),
		LogTypeName:        item.LogTypeName,
		DescriptionPreview: item.DescriptionPreview,
	}
}

func vaccinationAppListItemToDTO(item domainmodel.VaccinationListItem) dto.VaccinationListItemResponse {
	return dto.VaccinationListItemResponse{
		ID:                  item.ID.String(),
		PetID:               item.PetID.String(),
		GeneratedFromID:     uuidStringPtr(item.GeneratedFromID),
		Status:              item.Status,
		VaccineName:         item.VaccineName,
		CatalogMedicationID: uuidStringPtr(item.CatalogMedicationID),
		Targets:             healthDictionaryItemsToDTO(item.Targets),
		ScheduledAt:         timeStringPtr(item.ScheduledAt),
		AdministeredAt:      timeStringPtr(item.AdministeredAt),
		NextDueAt:           timeStringPtr(item.NextDueAt),
		VetVisitID:          uuidStringPtr(item.VetVisitID),
		ClinicName:          item.ClinicName,
		VetName:             item.VetName,
		NotesPreview:        item.NotesPreview,
		AttachmentsCount:    item.AttachmentsCount,
		RowVersion:          item.RowVersion,
		CreatedAt:           item.CreatedAt.UTC().Format(time.RFC3339),
		CreatedByUserID:     item.CreatedByUserID.String(),
		UpdatedAt:           item.UpdatedAt.UTC().Format(time.RFC3339),
		UpdatedByUserID:     item.UpdatedByUserID.String(),
	}
}

func vaccinationAppToDTO(item *domainmodel.Vaccination) dto.VaccinationResponse {
	attachments := make([]dto.HealthAttachmentResponse, 0, len(item.Attachments))
	for i := range item.Attachments {
		attachments = append(attachments, healthAttachmentAppToDTO(item.Attachments[i]))
	}
	return dto.VaccinationResponse{
		ID:                  item.ID.String(),
		PetID:               item.PetID.String(),
		GeneratedFromID:     uuidStringPtr(item.GeneratedFromID),
		Status:              item.Status,
		VaccineName:         item.VaccineName,
		CatalogMedicationID: uuidStringPtr(item.CatalogMedicationID),
		Targets:             healthDictionaryItemsToDTO(item.Targets),
		ScheduledAt:         timeStringPtr(item.ScheduledAt),
		AdministeredAt:      timeStringPtr(item.AdministeredAt),
		NextDueAt:           timeStringPtr(item.NextDueAt),
		VetVisitID:          uuidStringPtr(item.VetVisitID),
		ClinicName:          item.ClinicName,
		VetName:             item.VetName,
		Notes:               item.Notes,
		Attachments:         attachments,
		RowVersion:          item.RowVersion,
		CreatedAt:           item.CreatedAt.UTC().Format(time.RFC3339),
		CreatedByUserID:     item.CreatedByUserID.String(),
		UpdatedAt:           item.UpdatedAt.UTC().Format(time.RFC3339),
		UpdatedByUserID:     item.UpdatedByUserID.String(),
		CanEdit:             true,
		CanDelete:           true,
	}
}

func procedureAppListItemToDTO(item domainmodel.ProcedureListItem) dto.ProcedureListItemResponse {
	return dto.ProcedureListItemResponse{
		ID:                  item.ID.String(),
		PetID:               item.PetID.String(),
		GeneratedFromID:     uuidStringPtr(item.GeneratedFromID),
		Status:              item.Status,
		ProcedureTypeItem:   healthDictionaryItemPtrToDTO(item.ProcedureTypeItem),
		Title:               item.Title,
		DescriptionPreview:  item.DescriptionPreview,
		CatalogMedicationID: uuidStringPtr(item.CatalogMedicationID),
		ProductName:         item.ProductName,
		ScheduledAt:         timeStringPtr(item.ScheduledAt),
		PerformedAt:         timeStringPtr(item.PerformedAt),
		NextDueAt:           timeStringPtr(item.NextDueAt),
		VetVisitID:          uuidStringPtr(item.VetVisitID),
		NotesPreview:        item.NotesPreview,
		AttachmentsCount:    item.AttachmentsCount,
		RowVersion:          item.RowVersion,
		CreatedAt:           item.CreatedAt.UTC().Format(time.RFC3339),
		CreatedByUserID:     item.CreatedByUserID.String(),
		UpdatedAt:           item.UpdatedAt.UTC().Format(time.RFC3339),
		UpdatedByUserID:     item.UpdatedByUserID.String(),
	}
}

func procedureAppToDTO(item *domainmodel.Procedure) dto.ProcedureResponse {
	attachments := make([]dto.HealthAttachmentResponse, 0, len(item.Attachments))
	for i := range item.Attachments {
		attachments = append(attachments, healthAttachmentAppToDTO(item.Attachments[i]))
	}
	return dto.ProcedureResponse{
		ID:                  item.ID.String(),
		PetID:               item.PetID.String(),
		GeneratedFromID:     uuidStringPtr(item.GeneratedFromID),
		Status:              item.Status,
		ProcedureTypeItem:   healthDictionaryItemPtrToDTO(item.ProcedureTypeItem),
		Title:               item.Title,
		Description:         item.Description,
		CatalogMedicationID: uuidStringPtr(item.CatalogMedicationID),
		ProductName:         item.ProductName,
		ScheduledAt:         timeStringPtr(item.ScheduledAt),
		PerformedAt:         timeStringPtr(item.PerformedAt),
		NextDueAt:           timeStringPtr(item.NextDueAt),
		VetVisitID:          uuidStringPtr(item.VetVisitID),
		Notes:               item.Notes,
		Attachments:         attachments,
		RowVersion:          item.RowVersion,
		CreatedAt:           item.CreatedAt.UTC().Format(time.RFC3339),
		CreatedByUserID:     item.CreatedByUserID.String(),
		UpdatedAt:           item.UpdatedAt.UTC().Format(time.RFC3339),
		UpdatedByUserID:     item.UpdatedByUserID.String(),
		CanEdit:             true,
		CanDelete:           true,
	}
}

func medicalRecordAppListItemToDTO(item domainmodel.MedicalRecordListItem) dto.MedicalRecordListItemResponse {
	return dto.MedicalRecordListItemResponse{
		ID:                 item.ID.String(),
		PetID:              item.PetID.String(),
		RecordTypeItem:     healthDictionaryItemPtrToDTO(item.RecordTypeItem),
		Status:             item.Status,
		Title:              item.Title,
		DescriptionPreview: item.DescriptionPreview,
		StartedAt:          timeStringPtr(item.StartedAt),
		ResolvedAt:         timeStringPtr(item.ResolvedAt),
		AttachmentsCount:   item.AttachmentsCount,
		RowVersion:         item.RowVersion,
		CreatedAt:          item.CreatedAt.UTC().Format(time.RFC3339),
		CreatedByUserID:    item.CreatedByUserID.String(),
		UpdatedAt:          item.UpdatedAt.UTC().Format(time.RFC3339),
		UpdatedByUserID:    item.UpdatedByUserID.String(),
	}
}

func medicalRecordAppToDTO(item *domainmodel.MedicalRecord) dto.MedicalRecordResponse {
	attachments := make([]dto.HealthAttachmentResponse, 0, len(item.Attachments))
	for i := range item.Attachments {
		attachments = append(attachments, healthAttachmentAppToDTO(item.Attachments[i]))
	}
	return dto.MedicalRecordResponse{
		ID:             item.ID.String(),
		PetID:          item.PetID.String(),
		RecordTypeItem: healthDictionaryItemPtrToDTO(item.RecordTypeItem),
		Status:         item.Status,
		Title:          item.Title,
		Description:    item.Description,
		StartedAt:      timeStringPtr(item.StartedAt),
		ResolvedAt:     timeStringPtr(item.ResolvedAt),
		Attachments:    attachments,
		RowVersion:     item.RowVersion,
		CreatedAt:      item.CreatedAt.UTC().Format(time.RFC3339),
		CreatedByUser:  item.CreatedByUserID.String(),
		UpdatedAt:      item.UpdatedAt.UTC().Format(time.RFC3339),
		UpdatedByUser:  item.UpdatedByUserID.String(),
		CanEdit:        true,
		CanDelete:      true,
	}
}

func petDocumentAppToDTO(item domainmodel.PetDocument) dto.PetDocumentResponse {
	return dto.PetDocumentResponse{
		ID:            item.ID.String(),
		FileID:        item.FileID.String(),
		FileName:      item.FileName,
		FileType:      item.FileType,
		DownloadURL:   item.DownloadURL,
		PreviewURL:    item.PreviewURL,
		AddedAt:       item.AddedAt.UTC().Format(time.RFC3339),
		AddedByUserID: item.AddedByUserID.String(),
		EntityType:    item.EntityType,
		EntityID:      item.EntityID.String(),
	}
}

func scheduledItemListItemAppToDTO(item domainmodel.ScheduledItemListItem) dto.ScheduledItemListItemResponse {
	return dto.ScheduledItemListItemResponse{
		ID:                  item.ID.String(),
		PetID:               item.PetID.String(),
		SourceType:          item.SourceType,
		SourceID:            uuidStringPtr(item.SourceID),
		Title:               item.Title,
		NotePreview:         item.NotePreview,
		StartsAt:            item.StartsAt.UTC().Format(time.RFC3339),
		PushEnabled:         boolPtr(item.PushEnabled),
		RemindOffsetMinutes: item.RemindOffsetMinutes,
		Recurrence:          recurrenceToDTO(item.RecurrenceRule, item.RecurrenceInterval, item.RecurrenceUntil),
		RowVersion:          item.RowVersion,
		CreatedAt:           item.CreatedAt.UTC().Format(time.RFC3339),
		CreatedByUserID:     item.CreatedByUserID.String(),
		UpdatedAt:           item.UpdatedAt.UTC().Format(time.RFC3339),
		UpdatedByUserID:     item.UpdatedByUserID.String(),
	}
}

func scheduledItemAppToDTO(item *domainmodel.ScheduledItem) dto.ScheduledItemResponse {
	return dto.ScheduledItemResponse{
		ID:                  item.ID.String(),
		PetID:               item.PetID.String(),
		SourceType:          item.SourceType,
		SourceID:            uuidStringPtr(item.SourceID),
		Title:               item.Title,
		Note:                item.Note,
		StartsAt:            item.StartsAt.UTC().Format(time.RFC3339),
		PushEnabled:         boolPtr(item.PushEnabled),
		RemindOffsetMinutes: item.RemindOffsetMinutes,
		Recurrence:          recurrenceToDTO(item.RecurrenceRule, item.RecurrenceInterval, item.RecurrenceUntil),
		RowVersion:          item.RowVersion,
		CreatedAt:           item.CreatedAt.UTC().Format(time.RFC3339),
		CreatedByUserID:     item.CreatedByUserID.String(),
		UpdatedAt:           item.UpdatedAt.UTC().Format(time.RFC3339),
		UpdatedByUserID:     item.UpdatedByUserID.String(),
	}
}

func scheduledOccurrenceAppToDTO(item domainmodel.ScheduledItemOccurrenceListItem) dto.ScheduledOccurrenceResponse {
	return dto.ScheduledOccurrenceResponse{
		ID:              item.ID.String(),
		ScheduledItemID: item.ScheduledItemID.String(),
		PetID:           item.PetID.String(),
		ScheduledFor:    item.ScheduledFor.UTC().Format(time.RFC3339),
		CreatedAt:       item.CreatedAt.UTC().Format(time.RFC3339),
		Rule:            scheduledItemAppToDTO(&item.Rule),
	}
}

func calendarDayItemAppToDTO(item domainmodel.CalendarDayItem) dto.CalendarDayItemResponse {
	return dto.CalendarDayItemResponse{
		ItemType:              item.ItemType,
		EntityID:              item.EntityID.String(),
		PetID:                 item.PetID.String(),
		Title:                 item.Title,
		Subtitle:              item.Subtitle,
		ScheduledFor:          item.ScheduledFor.UTC().Format(time.RFC3339),
		Status:                item.Status,
		VisitID:               uuidStringPtr(item.VisitID),
		VaccinationID:         uuidStringPtr(item.VaccinationID),
		ProcedureID:           uuidStringPtr(item.ProcedureID),
		ScheduledItemID:       uuidStringPtr(item.ScheduledItemID),
		ScheduledOccurrenceID: uuidStringPtr(item.ScheduledOccurrenceID),
	}
}

func calendarDateMarkerAppToDTO(item domainmodel.CalendarDateMarker) dto.CalendarMarkerResponse {
	return dto.CalendarMarkerResponse{
		Date:           item.Date.UTC().Format("2006-01-02"),
		PlannedCount:   item.PlannedCount,
		CompletedCount: item.CompletedCount,
		TotalCount:     item.TotalCount,
	}
}

func recurrenceToDTO(rule *string, interval *int, until *time.Time) *dto.RecurrenceResponse {
	if rule == nil {
		return nil
	}
	return &dto.RecurrenceResponse{
		Rule:     *rule,
		Interval: interval,
		Until:    timeStringPtr(until),
	}
}

func uuidStringPtr(v *uuid.UUID) *string {
	if v == nil {
		return nil
	}
	s := v.String()
	return &s
}

func timeStringPtr(v *time.Time) *string {
	if v == nil {
		return nil
	}
	s := v.UTC().Format(time.RFC3339)
	return &s
}

func boolPtr(v bool) *bool {
	return &v
}
