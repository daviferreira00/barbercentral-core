package reports

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type Service interface {
	GetRevenueReport(ctx context.Context, clientID string, startDate, endDate string, professionalID, serviceID string) (*RevenueReport, error)
	GetOccupancyReport(ctx context.Context, clientID string, startDate, endDate string, professionalID string) (*OccupancyReport, error)
	GetCustomerReport(ctx context.Context, clientID string, startDate, endDate string) (*CustomerReport, error)
	GetCancellationReport(ctx context.Context, clientID string, startDate, endDate string) (*CancellationReport, error)

	ExportRevenueCSV(ctx context.Context, clientID string, startDate, endDate string) ([]byte, error)
	ExportOccupancyCSV(ctx context.Context, clientID string, startDate, endDate string) ([]byte, error)
	ExportCustomersCSV(ctx context.Context, clientID string, startDate, endDate string) ([]byte, error)
	ExportCancellationsCSV(ctx context.Context, clientID string, startDate, endDate string) ([]byte, error)

	ExportRevenuePDF(ctx context.Context, clientID string, startDate, endDate string) ([]byte, error)
}

type RevenueReport struct {
	TotalRevenue    float64      `json:"total_revenue"`
	AverageTicket   float64      `json:"average_ticket"`
	TotalPayments   int          `json:"total_payments"`
	DailyRevenue    []ChartItem  `json:"daily_revenue"`
	ProfessionalRev []ReportItem `json:"professional_revenue"`
	ServiceRev      []ReportItem `json:"service_revenue"`
	MethodRev       []ChartItem  `json:"method_revenue"`
}

type OccupancyReport struct {
	OccupancyRate   float64           `json:"occupancy_rate"`
	ProfessionalOcc []ReportItem      `json:"professional_occupancy"`
	WeekdayHeatmap  []HeatmapItem     `json:"weekday_heatmap"`
	ProfessionalHrs []ProfessionalHrs `json:"professional_hours"`
}

type CustomerReport struct {
	TotalCustomers int          `json:"total_customers"`
	NewCustomers   int          `json:"new_customers"`
	ReturningCust  int          `json:"returning_customers"`
	ReturnRate     float64      `json:"return_rate"`
	DailyNewCust   []ChartItem  `json:"daily_new_customers"`
	TopCustomers   []ReportItem `json:"top_customers"`
	ChurnCustomers []ReportItem `json:"churn_customers"`
}

type CancellationReport struct {
	TotalAppointments int            `json:"total_appointments"`
	Cancellations     int            `json:"cancellations"`
	NoShows           int            `json:"no_shows"`
	CancellationRate  float64        `json:"cancellation_rate"`
	DailyCancels      []ChartItem    `json:"daily_cancellations"`
	ProfessionalCanc  []ReportItem   `json:"professional_cancellations"`
	RecentCancels     []RecentCancel `json:"recent_cancellations"`
}

type ChartItem struct {
	Label string  `db:"label" json:"label"`
	Value float64 `db:"value" json:"value"`
}

type ReportItem struct {
	Label string  `db:"label" json:"label"`
	Count int     `db:"count" json:"count"`
	Value float64 `db:"value" json:"value"`
	Date  *string `db:"date" json:"date,omitempty"`
}

type HeatmapItem struct {
	Weekday int `db:"weekday" json:"weekday"`
	Hour    int `db:"hour" json:"hour"`
	Count   int `db:"count" json:"count"`
}

type ProfessionalHrs struct {
	Label        string  `db:"label" json:"label"`
	HoursAllowed float64 `json:"hours_allowed"`
	HoursBooked  float64 `db:"hours_booked" json:"hours_booked"`
	Percent      float64 `json:"percent"`
	Count        int     `db:"count" json:"count"`
}

type RecentCancel struct {
	Date   string  `db:"label" json:"date"`
	Time   string  `db:"time" json:"time"`
	Prof   string  `db:"prof" json:"professional"`
	Cust   *string `db:"cust" json:"customer,omitempty"`
	Reason *string `db:"reason" json:"reason,omitempty"`
}

type service struct {
	db *sqlx.DB
}

func NewService(db *sqlx.DB) Service {
	return &service{db: db}
}

func (s *service) GetRevenueReport(ctx context.Context, clientID string, startDate, endDate string, professionalID, serviceID string) (*RevenueReport, error) {
	// 1. Resumo Geral
	var summary struct {
		TotalRevenue  float64 `db:"total_revenue"`
		AverageTicket float64 `db:"average_ticket"`
		TotalPayments int     `db:"total_count"`
	}

	querySum := `SELECT COALESCE(SUM(ap.amount), 0) AS total_revenue, COALESCE(AVG(ap.amount), 0) AS average_ticket, COUNT(ap.id) AS total_count 
	             FROM appointment_payment ap 
	             JOIN appointment a ON ap.appointment_id = a.id 
	             WHERE ap.client_id = ? AND ap.status = 'paid' AND a.date BETWEEN ? AND ?`

	var args []interface{}
	args = append(args, clientID, startDate, endDate)

	if professionalID != "" {
		querySum += " AND a.professional_id = ?"
		args = append(args, professionalID)
	}
	if serviceID != "" {
		querySum += " AND EXISTS (SELECT 1 FROM appointment_service WHERE appointment_id = a.id AND service_id = ?)"
		args = append(args, serviceID)
	}

	err := s.db.GetContext(ctx, &summary, querySum, args...)
	if err != nil {
		return nil, err
	}

	// 2. Faturamento por Dia
	var daily []ChartItem
	queryDaily := `SELECT a.date AS label, COALESCE(SUM(ap.amount), 0) AS value 
	               FROM appointment_payment ap 
	               JOIN appointment a ON ap.appointment_id = a.id 
	               WHERE ap.client_id = ? AND ap.status = 'paid' AND a.date BETWEEN ? AND ?`
	var argsDaily []interface{}
	argsDaily = append(argsDaily, clientID, startDate, endDate)

	if professionalID != "" {
		queryDaily += " AND a.professional_id = ?"
		argsDaily = append(argsDaily, professionalID)
	}
	if serviceID != "" {
		queryDaily += " AND EXISTS (SELECT 1 FROM appointment_service WHERE appointment_id = a.id AND service_id = ?)"
		argsDaily = append(argsDaily, serviceID)
	}
	queryDaily += " GROUP BY a.date ORDER BY a.date ASC"

	err = s.db.SelectContext(ctx, &daily, queryDaily, argsDaily...)
	if err != nil {
		return nil, err
	}

	// 3. Faturamento por Profissional
	var prof []ReportItem
	queryProf := `SELECT p.name AS label, COUNT(a.id) AS count, COALESCE(SUM(ap.amount), 0) AS value 
	              FROM appointment_payment ap 
	              JOIN appointment a ON ap.appointment_id = a.id 
	              JOIN professional p ON a.professional_id = p.id 
	              WHERE ap.client_id = ? AND ap.status = 'paid' AND a.date BETWEEN ? AND ? 
	              GROUP BY p.name`
	err = s.db.SelectContext(ctx, &prof, queryProf, clientID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 4. Faturamento por Serviço
	var svc []ReportItem
	querySvc := `SELECT s.name AS label, COUNT(aps.id) AS count, COALESCE(SUM(aps.price), 0) AS value 
	             FROM appointment_payment ap 
	             JOIN appointment a ON ap.appointment_id = a.id 
	             JOIN appointment_service aps ON a.id = aps.appointment_id 
	             JOIN service s ON aps.service_id = s.id 
	             WHERE ap.client_id = ? AND ap.status = 'paid' AND a.date BETWEEN ? AND ? 
	             GROUP BY s.name`
	err = s.db.SelectContext(ctx, &svc, querySvc, clientID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 5. Método de Pagamento
	var methods []ChartItem
	queryMethod := `SELECT ap.method AS label, COALESCE(SUM(ap.amount), 0) AS value 
	                FROM appointment_payment ap 
	                JOIN appointment a ON ap.appointment_id = a.id 
	                WHERE ap.client_id = ? AND ap.status = 'paid' AND a.date BETWEEN ? AND ? 
	                GROUP BY ap.method`
	err = s.db.SelectContext(ctx, &methods, queryMethod, clientID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	return &RevenueReport{
		TotalRevenue:    summary.TotalRevenue,
		AverageTicket:   summary.AverageTicket,
		TotalPayments:   summary.TotalPayments,
		DailyRevenue:    daily,
		ProfessionalRev: prof,
		ServiceRev:      svc,
		MethodRev:       methods,
	}, nil
}

func (s *service) GetOccupancyReport(ctx context.Context, clientID string, startDate, endDate string, professionalID string) (*OccupancyReport, error) {
	// 1. Horas agendadas (soma de durações de serviços concluídos ou confirmados)
	var totalMinutesBooked float64
	queryBooked := `SELECT COALESCE(SUM(aps.duration_minutes), 0) 
	                FROM appointment_service aps 
	                JOIN appointment a ON aps.appointment_id = a.id 
	                WHERE a.client_id = ? AND a.status IN ('confirmed', 'completed') AND a.date BETWEEN ? AND ?`
	var argsBooked []interface{}
	argsBooked = append(argsBooked, clientID, startDate, endDate)
	if professionalID != "" {
		queryBooked += " AND a.professional_id = ?"
		argsBooked = append(argsBooked, professionalID)
	}
	err := s.db.GetContext(ctx, &totalMinutesBooked, queryBooked, argsBooked...)
	if err != nil {
		return nil, err
	}

	// 2. Profissionais cadastrados ativos
	var countProfs float64
	queryCount := "SELECT COUNT(*) FROM professional WHERE client_id = ? AND active = 1"
	var argsCount []interface{}
	argsCount = append(argsCount, clientID)
	if professionalID != "" {
		queryCount += " AND id = ?"
		argsCount = append(argsCount, professionalID)
	}
	err = s.db.GetContext(ctx, &countProfs, queryCount, argsCount...)
	if err != nil {
		return nil, err
	}

	if countProfs == 0 {
		countProfs = 1.0
	}

	// Calcula dias úteis no período
	start, _ := time.Parse("2006-01-02", startDate)
	end, _ := time.Parse("2006-01-02", endDate)
	days := float64(int(end.Sub(start).Hours()/24) + 1)
	if days <= 0 {
		days = 1.0
	}

	// 8 horas por dia por profissional = 480 minutos
	totalMinutesAllowed := countProfs * days * 480.0
	occupancyRate := (totalMinutesBooked / totalMinutesAllowed) * 100.0
	if occupancyRate > 100 {
		occupancyRate = 100.0
	}

	// 3. Ocupação por Profissional (gráfico horizontal)
	var listProfs []ProfessionalHrs
	queryProfHrs := `SELECT p.name AS label, COUNT(a.id) AS count, COALESCE(SUM(aps.duration_minutes), 0) AS hours_booked 
	                 FROM appointment a 
	                 JOIN professional p ON a.professional_id = p.id 
	                 JOIN appointment_service aps ON a.id = aps.appointment_id 
	                 WHERE a.client_id = ? AND a.status IN ('confirmed', 'completed') AND a.date BETWEEN ? AND ? 
	                 GROUP BY p.name`
	err = s.db.SelectContext(ctx, &listProfs, queryProfHrs, clientID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	for i := range listProfs {
		// Converte minutos agendados para horas agendadas
		listProfs[i].HoursBooked = listProfs[i].HoursBooked / 60.0
		listProfs[i].HoursAllowed = days * 8.0 // 8 horas por dia
		listProfs[i].Percent = (listProfs[i].HoursBooked / listProfs[i].HoursAllowed) * 100.0
		if listProfs[i].Percent > 100 {
			listProfs[i].Percent = 100.0
		}
	}

	// 4. Heatmap semanal
	var heatmap []HeatmapItem
	queryHeat := `SELECT DAYOFWEEK(a.date) AS weekday, HOUR(a.start_time) AS hour, COUNT(*) AS count 
	              FROM appointment a 
	              WHERE a.client_id = ? AND a.status IN ('confirmed', 'completed') AND a.date BETWEEN ? AND ? 
	              GROUP BY weekday, hour`
	err = s.db.SelectContext(ctx, &heatmap, queryHeat, clientID, startDate, endDate)
	if err != nil {
		// Ignora erros menores e retorna heatmap vazio se der problema de timezone/hour parsing
		heatmap = []HeatmapItem{}
	}

	return &OccupancyReport{
		OccupancyRate:   occupancyRate,
		ProfessionalOcc: []ReportItem{}, // Usar professionalHrs diretamente
		WeekdayHeatmap:  heatmap,
		ProfessionalHrs: listProfs,
	}, nil
}

func (s *service) GetCustomerReport(ctx context.Context, clientID string, startDate, endDate string) (*CustomerReport, error) {
	// 1. Total Clientes
	var totalCust int
	err := s.db.GetContext(ctx, &totalCust, "SELECT COUNT(*) FROM customer WHERE client_id = ?", clientID)
	if err != nil {
		return nil, err
	}

	// 2. Novos Clientes no período
	var newCust int
	queryNew := `SELECT COUNT(*) FROM (
	                SELECT customer_id, MIN(date) AS first_date 
	                FROM appointment 
	                WHERE client_id = ? AND customer_id IS NOT NULL AND status IN ('confirmed', 'completed') 
	                GROUP BY customer_id
	             ) AS sub WHERE first_date BETWEEN ? AND ?`
	err = s.db.GetContext(ctx, &newCust, queryNew, clientID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 3. Clientes recorrentes (que fizeram agendamento no período e já haviam feito antes)
	var returningCust int
	queryRet := `SELECT COUNT(*) FROM (
	                SELECT customer_id FROM appointment 
	                WHERE client_id = ? AND customer_id IS NOT NULL AND status IN ('confirmed', 'completed') AND date BETWEEN ? AND ?
	                  AND customer_id IN (SELECT customer_id FROM appointment WHERE client_id = ? AND date < ? AND status IN ('confirmed', 'completed')) 
	                GROUP BY customer_id
	             ) AS sub`
	err = s.db.GetContext(ctx, &returningCust, queryRet, clientID, startDate, endDate, clientID, startDate)
	if err != nil {
		return nil, err
	}

	// Taxa de retorno
	var returnRate float64
	totalActiveInPeriod := newCust + returningCust
	if totalActiveInPeriod > 0 {
		returnRate = (float64(returningCust) / float64(totalActiveInPeriod)) * 100.0
	}

	// 4. Novos clientes ao longo do tempo
	var dailyNew []ChartItem
	queryDailyNew := `SELECT first_date AS label, COUNT(*) AS value FROM (
	                     SELECT customer_id, MIN(date) AS first_date 
	                     FROM appointment 
	                     WHERE client_id = ? AND customer_id IS NOT NULL AND status IN ('confirmed', 'completed') 
	                     GROUP BY customer_id
	                  ) AS sub WHERE first_date BETWEEN ? AND ? GROUP BY first_date ORDER BY first_date ASC`
	err = s.db.SelectContext(ctx, &dailyNew, queryDailyNew, clientID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 5. Top 10 Clientes por valor gasto
	var topCustomers []ReportItem
	queryTop := `SELECT c.name AS label, COUNT(a.id) AS count, COALESCE(SUM(ap.amount), 0) AS value, MAX(a.date) AS date 
	             FROM appointment_payment ap 
	             JOIN appointment a ON ap.appointment_id = a.id 
	             JOIN customer c ON a.customer_id = c.id 
	             WHERE ap.client_id = ? AND ap.status = 'paid' AND a.date BETWEEN ? AND ? 
	             GROUP BY c.name 
	             ORDER BY value DESC 
	             LIMIT 10`
	err = s.db.SelectContext(ctx, &topCustomers, queryTop, clientID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 6. Churn (Clientes sem retorno nos últimos 60 dias)
	limitDate := time.Now().AddDate(0, 0, -60).Format("2006-01-02")
	var churn []ReportItem
	queryChurn := `SELECT c.name AS label, MAX(a.date) AS date 
	               FROM customer c 
	               JOIN appointment a ON a.customer_id = c.id 
	               WHERE c.client_id = ? AND a.status IN ('confirmed', 'completed') 
	               GROUP BY c.name 
	               HAVING MAX(a.date) < ? 
	               ORDER BY date DESC`
	err = s.db.SelectContext(ctx, &churn, queryChurn, clientID, limitDate)
	if err != nil {
		return nil, err
	}

	return &CustomerReport{
		TotalCustomers: totalCust,
		NewCustomers:   newCust,
		ReturningCust:  returningCust,
		ReturnRate:     returnRate,
		DailyNewCust:   dailyNew,
		TopCustomers:   topCustomers,
		ChurnCustomers: churn,
	}, nil
}

func (s *service) GetCancellationReport(ctx context.Context, clientID string, startDate, endDate string) (*CancellationReport, error) {
	// 1. Resumo
	var summary struct {
		Total   int `db:"total"`
		Cancels int `db:"cancels"`
		NoShows int `db:"noshows"`
	}

	querySum := `SELECT COUNT(*) AS total, 
	                    SUM(CASE WHEN status = 'cancelled' THEN 1 ELSE 0 END) AS cancels, 
	                    SUM(CASE WHEN status = 'no_show' THEN 1 ELSE 0 END) AS noshows 
	             FROM appointment 
	             WHERE client_id = ? AND date BETWEEN ? AND ?`
	err := s.db.GetContext(ctx, &summary, querySum, clientID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	rate := 0.0
	if summary.Total > 0 {
		rate = (float64(summary.Cancels+summary.NoShows) / float64(summary.Total)) * 100.0
	}

	// 2. Timeline de cancelamentos
	var daily []ChartItem
	queryDaily := `SELECT date AS label, COUNT(*) AS value 
	               FROM appointment 
	               WHERE client_id = ? AND status IN ('cancelled', 'no_show') AND date BETWEEN ? AND ? 
	               GROUP BY date ORDER BY date ASC`
	err = s.db.SelectContext(ctx, &daily, queryDaily, clientID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 3. Por Profissional
	var prof []ReportItem
	queryProf := `SELECT p.name AS label, COUNT(a.id) AS count 
	              FROM appointment a 
	              JOIN professional p ON a.professional_id = p.id 
	              WHERE a.client_id = ? AND a.status IN ('cancelled', 'no_show') AND a.date BETWEEN ? AND ? 
	              GROUP BY p.name`
	err = s.db.SelectContext(ctx, &prof, queryProf, clientID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 4. Lista recente de cancelamentos com motivo
	var recent []RecentCancel
	queryRecent := `SELECT a.date AS label, a.start_time AS time, p.name AS prof, c.name AS cust, asl.notes AS reason 
	                FROM appointment a 
	                JOIN professional p ON a.professional_id = p.id 
	                LEFT JOIN customer c ON a.customer_id = c.id 
	                LEFT JOIN appointment_status_log asl ON asl.appointment_id = a.id AND asl.to_status = 'cancelled' 
	                WHERE a.client_id = ? AND a.status = 'cancelled' AND a.date BETWEEN ? AND ? 
	                ORDER BY a.date DESC LIMIT 15`
	err = s.db.SelectContext(ctx, &recent, queryRecent, clientID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	return &CancellationReport{
		TotalAppointments: summary.Total,
		Cancellations:     summary.Cancels,
		NoShows:           summary.NoShows,
		CancellationRate:  rate,
		DailyCancels:      daily,
		ProfessionalCanc:  prof,
		RecentCancels:     recent,
	}, nil
}

// EXPORTAÇÕES CSV

func (s *service) ExportRevenueCSV(ctx context.Context, clientID string, startDate, endDate string) ([]byte, error) {
	rep, err := s.GetRevenueReport(ctx, clientID, startDate, endDate, "", "")
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	buf.Write([]byte{0xEF, 0xBB, 0xBF}) // BOM
	w := csv.NewWriter(&buf)
	w.Comma = ';'

	_ = w.Write([]string{"RESUMO GERAL DO FATURAMENTO"})
	_ = w.Write([]string{"Total Faturado", fmt.Sprintf("R$ %.2f", rep.TotalRevenue)})
	_ = w.Write([]string{"Ticket Médio", fmt.Sprintf("R$ %.2f", rep.AverageTicket)})
	_ = w.Write([]string{"Total Atendimentos Pagos", fmt.Sprintf("%d", rep.TotalPayments)})
	_ = w.Write([]string{""})

	_ = w.Write([]string{"FATURAMENTO POR PROFISSIONAL"})
	_ = w.Write([]string{"Profissional", "Atendimentos", "Valor Total"})
	for _, p := range rep.ProfessionalRev {
		_ = w.Write([]string{p.Label, fmt.Sprintf("%d", p.Count), fmt.Sprintf("R$ %.2f", p.Value)})
	}
	_ = w.Write([]string{""})

	_ = w.Write([]string{"FATURAMENTO POR SERVIÇO"})
	_ = w.Write([]string{"Serviço", "Quantidade", "Valor Total"})
	for _, sv := range rep.ServiceRev {
		_ = w.Write([]string{sv.Label, fmt.Sprintf("%d", sv.Count), fmt.Sprintf("R$ %.2f", sv.Value)})
	}

	w.Flush()
	return buf.Bytes(), nil
}

func (s *service) ExportOccupancyCSV(ctx context.Context, clientID string, startDate, endDate string) ([]byte, error) {
	rep, err := s.GetOccupancyReport(ctx, clientID, startDate, endDate, "")
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	buf.Write([]byte{0xEF, 0xBB, 0xBF})
	w := csv.NewWriter(&buf)
	w.Comma = ';'

	_ = w.Write([]string{"RELATÓRIO DE OCUPAÇÃO DE AGENDA"})
	_ = w.Write([]string{"Taxa de Ocupação Geral", fmt.Sprintf("%.2f%%", rep.OccupancyRate)})
	_ = w.Write([]string{""})

	_ = w.Write([]string{"OCUPAÇÃO DETALHADA POR PROFISSIONAL"})
	_ = w.Write([]string{"Profissional", "Horas Disponíveis", "Horas Agendadas", "Aproveitamento (%)", "Qtd Agendamentos"})
	for _, ph := range rep.ProfessionalHrs {
		_ = w.Write([]string{
			ph.Label,
			fmt.Sprintf("%.1fh", ph.HoursAllowed),
			fmt.Sprintf("%.1fh", ph.HoursBooked),
			fmt.Sprintf("%.2f%%", ph.Percent),
			fmt.Sprintf("%d", ph.Count),
		})
	}

	w.Flush()
	return buf.Bytes(), nil
}

func (s *service) ExportCustomersCSV(ctx context.Context, clientID string, startDate, endDate string) ([]byte, error) {
	rep, err := s.GetCustomerReport(ctx, clientID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	buf.Write([]byte{0xEF, 0xBB, 0xBF})
	w := csv.NewWriter(&buf)
	w.Comma = ';'

	_ = w.Write([]string{"ANÁLISE DE CLIENTES E RETENÇÃO"})
	_ = w.Write([]string{"Clientes Cadastrados", fmt.Sprintf("%d", rep.TotalCustomers)})
	_ = w.Write([]string{"Novos no Período", fmt.Sprintf("%d", rep.NewCustomers)})
	_ = w.Write([]string{"Recorrentes", fmt.Sprintf("%d", rep.ReturningCust)})
	_ = w.Write([]string{"Taxa de Retorno", fmt.Sprintf("%.2f%%", rep.ReturnRate)})
	_ = w.Write([]string{""})

	_ = w.Write([]string{"RANKING TOP 10 CLIENTES (POR VALOR GASTO)"})
	_ = w.Write([]string{"Cliente", "Visitas no Período", "Gasto Total", "Última Visita"})
	for _, c := range rep.TopCustomers {
		lastV := ""
		if c.Date != nil {
			lastV = *c.Date
		}
		_ = w.Write([]string{c.Label, fmt.Sprintf("%d", c.Count), fmt.Sprintf("R$ %.2f", c.Value), lastV})
	}

	w.Flush()
	return buf.Bytes(), nil
}

func (s *service) ExportCancellationsCSV(ctx context.Context, clientID string, startDate, endDate string) ([]byte, error) {
	rep, err := s.GetCancellationReport(ctx, clientID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	buf.Write([]byte{0xEF, 0xBB, 0xBF})
	w := csv.NewWriter(&buf)
	w.Comma = ';'

	_ = w.Write([]string{"RELATÓRIO DE ABSENTEÍSMO E CANCELAMENTOS"})
	_ = w.Write([]string{"Total Agendamentos no Período", fmt.Sprintf("%d", rep.TotalAppointments)})
	_ = w.Write([]string{"Cancelados", fmt.Sprintf("%d", rep.Cancellations)})
	_ = w.Write([]string{"No-shows (Faltas)", fmt.Sprintf("%d", rep.NoShows)})
	_ = w.Write([]string{"Taxa Geral de Absenteísmo", fmt.Sprintf("%.2f%%", rep.CancellationRate)})
	_ = w.Write([]string{""})

	_ = w.Write([]string{"CANCELAMENTOS RECENTES"})
	_ = w.Write([]string{"Data", "Horário", "Profissional", "Cliente", "Motivo Informado"})
	for _, rc := range rep.RecentCancels {
		custStr := "Sem cadastro"
		if rc.Cust != nil {
			custStr = *rc.Cust
		}
		reasonStr := "Não informado"
		if rc.Reason != nil {
			reasonStr = *rc.Reason
		}
		_ = w.Write([]string{rc.Date, rc.Time, rc.Prof, custStr, reasonStr})
	}

	w.Flush()
	return buf.Bytes(), nil
}

// EXPORTAÇÃO PDF

func (s *service) ExportRevenuePDF(ctx context.Context, clientID string, startDate, endDate string) ([]byte, error) {
	rep, err := s.GetRevenueReport(ctx, clientID, startDate, endDate, "", "")
	if err != nil {
		return nil, err
	}

	// Gera um relatório HTML estruturado com CSS de impressão profissional.
	// O navegador/front faz o download direto ou abre o diálogo nativo do sistema.
	htmlContent := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<title>Relatório Financeiro de Faturamento</title>
	<style>
		body { font-family: sans-serif; font-size: 13px; color: #333; line-height: 1.5; margin: 30px; }
		h1 { font-size: 20px; font-weight: 800; border-bottom: 2px solid #333; padding-bottom: 8px; margin-bottom: 25px; text-transform: uppercase; }
		h2 { font-size: 14px; font-weight: bold; margin-top: 30px; border-bottom: 1px solid #ddd; padding-bottom: 4px; text-transform: uppercase; color: #555; }
		.summary-grid { display: grid; grid-template-cols: repeat(3, 1fr); gap: 15px; margin-bottom: 20px; }
		.summary-card { border: 1px solid #e2e8f0; border-radius: 8px; padding: 15px; background: #f8fafc; }
		.summary-label { font-size: 9px; font-weight: bold; color: #64748b; text-transform: uppercase; }
		.summary-val { font-size: 18px; font-weight: 900; color: #0f172a; margin-top: 5px; }
		table { width: 100%%; border-collapse: collapse; margin-top: 10px; }
		th { font-size: 10px; text-transform: uppercase; color: #64748b; background: #f8fafc; text-align: left; padding: 8px 12px; border-bottom: 1px solid #e2e8f0; }
		td { padding: 10px 12px; border-bottom: 1px solid #f1f5f9; }
		.text-right { text-align: right; }
		.font-bold { font-weight: bold; }
		@media print {
			body { margin: 15mm; }
			.no-print { display: none; }
		}
	</style>
</head>
<body onload="window.print()">
	<div class="no-print" style="background: #f1f5f9; padding: 10px 20px; border-radius: 6px; margin-bottom: 20px; display: flex; justify-content: space-between; align-items: center;">
		<span>Este relatório está formatado para impressão direta em formato PDF.</span>
		<button onclick="window.print()" style="font-weight: bold; cursor: pointer; padding: 6px 12px; background: #0f172a; color: white; border: none; border-radius: 4px;">Imprimir / Salvar PDF</button>
	</div>

	<h1>Relatório Financeiro — Período: %s a %s</h1>

	<div class="summary-grid">
		<div class="summary-card">
			<div class="summary-label">Total Faturado</div>
			<div class="summary-val">R$ %.2f</div>
		</div>
		<div class="summary-card">
			<div class="summary-label">Ticket Médio</div>
			<div class="summary-val">R$ %.2f</div>
		</div>
		<div class="summary-card">
			<div class="summary-label">Atendimentos Pagos</div>
			<div class="summary-val">%d</div>
		</div>
	</div>

	<h2>Faturamento por Profissional</h2>
	<table>
		<thead>
			<tr>
				<th>Profissional</th>
				<th class="text-right">Atendimentos</th>
				<th class="text-right">Total Faturado</th>
			</tr>
		</thead>
		<tbody>`,
		startDate, endDate, rep.TotalRevenue, rep.AverageTicket, rep.TotalPayments,
	)

	for _, p := range rep.ProfessionalRev {
		htmlContent += fmt.Sprintf(`
			<tr>
				<td class="font-bold">%s</td>
				<td class="text-right">%d</td>
				<td class="text-right font-bold">R$ %.2f</td>
			</tr>`,
			p.Label, p.Count, p.Value,
		)
	}

	htmlContent += `
		</tbody>
	</table>

	<h2>Faturamento por Serviço</h2>
	<table>
		<thead>
			<tr>
				<th>Serviço</th>
				<th class="text-right">Quantidade</th>
				<th class="text-right">Total Faturado</th>
			</tr>
		</thead>
		<tbody>`

	for _, sv := range rep.ServiceRev {
		htmlContent += fmt.Sprintf(`
			<tr>
				<td class="font-bold">%s</td>
				<td class="text-right">%d</td>
				<td class="text-right font-bold">R$ %.2f</td>
			</tr>`,
			sv.Label, sv.Count, sv.Value,
		)
	}

	htmlContent += `
		</tbody>
	</table>
</body>
</html>`

	return []byte(htmlContent), nil
}
