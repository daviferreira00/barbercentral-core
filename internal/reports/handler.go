package reports

import (
	"net/http"

	"barbercentral-core/internal/shared"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetRevenueReport(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)

	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	professionalID := r.URL.Query().Get("professional_id")
	serviceID := r.URL.Query().Get("service_id")

	if startDate == "" || endDate == "" {
		shared.RespondWithError(w, http.StatusBadRequest, "Parâmetros start_date e end_date são obrigatórios", nil)
		return
	}

	rep, err := h.service.GetRevenueReport(r.Context(), clientID, startDate, endDate, professionalID, serviceID)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao gerar relatório financeiro", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, rep)
}

func (h *Handler) GetOccupancyReport(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)

	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	professionalID := r.URL.Query().Get("professional_id")

	if startDate == "" || endDate == "" {
		shared.RespondWithError(w, http.StatusBadRequest, "Parâmetros start_date e end_date são obrigatórios", nil)
		return
	}

	rep, err := h.service.GetOccupancyReport(r.Context(), clientID, startDate, endDate, professionalID)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao gerar relatório de ocupação", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, rep)
}

func (h *Handler) GetCustomerReport(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)

	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	if startDate == "" || endDate == "" {
		shared.RespondWithError(w, http.StatusBadRequest, "Parâmetros start_date e end_date são obrigatórios", nil)
		return
	}

	rep, err := h.service.GetCustomerReport(r.Context(), clientID, startDate, endDate)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao gerar relatório de clientes", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, rep)
}

func (h *Handler) GetCancellationReport(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)

	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	if startDate == "" || endDate == "" {
		shared.RespondWithError(w, http.StatusBadRequest, "Parâmetros start_date e end_date são obrigatórios", nil)
		return
	}

	rep, err := h.service.GetCancellationReport(r.Context(), clientID, startDate, endDate)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao gerar relatório de cancelamentos", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, rep)
}

// EXPORTAÇÕES CSV

func (h *Handler) ExportRevenueCSV(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	csvData, err := h.service.ExportRevenueCSV(r.Context(), clientID, startDate, endDate)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao exportar CSV de faturamento", err)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=faturamento.csv")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(csvData)
}

func (h *Handler) ExportOccupancyCSV(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	csvData, err := h.service.ExportOccupancyCSV(r.Context(), clientID, startDate, endDate)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao exportar CSV de ocupação", err)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=ocupacao.csv")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(csvData)
}

func (h *Handler) ExportCustomersCSV(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	csvData, err := h.service.ExportCustomersCSV(r.Context(), clientID, startDate, endDate)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao exportar CSV de clientes", err)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=clientes.csv")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(csvData)
}

func (h *Handler) ExportCancellationsCSV(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	csvData, err := h.service.ExportCancellationsCSV(r.Context(), clientID, startDate, endDate)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao exportar CSV de cancelamentos", err)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=cancelamentos.csv")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(csvData)
}

// EXPORTAÇÃO PDF

func (h *Handler) ExportRevenuePDF(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	htmlData, err := h.service.ExportRevenuePDF(r.Context(), clientID, startDate, endDate)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao exportar PDF de faturamento", err)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(htmlData)
}
