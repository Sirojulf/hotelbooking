// Package ai - Agentic booking orchestrator menggunakan OpenAI Function Calling.
//
// Flow:
//   User message ──▶ OpenAI ──▶ tool_calls? ──▶ execute tools ──▶ feed back ──▶ loop
//                                    │
//                                    └─ no ──▶ final text answer to user
//
// Tools yang di-expose ke AI:
//   1. search_hotels      → cari hotel by name/city
//   2. list_rooms         → daftar kamar di hotel + price
//   3. check_availability → cek + dapatkan quote price
//   4. create_reservation → buat reservation (pending)
//   5. pay_reservation    → bayar reservation
package ai

import (
	"context"
	"fmt"
	"strings"
	"time"

	"hotelbooking/internal/models"
	"hotelbooking/internal/repository"
	"hotelbooking/internal/service"

	json "github.com/goccy/go-json"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
	"github.com/spf13/viper"
)

const (
	defaultModel  = openai.ChatModelGPT4oMini
	maxIterations = 10
	systemPrompt  = `Kamu adalah asisten booking hotel yang membantu user melakukan reservasi secara otomatis.

Hari ini: %s

Alur kerja utama (LAKUKAN SECARA OTOMATIS sebanyak mungkin tanpa banyak tanya):
1. Pahami kebutuhan user dari pesan (tanggal, hotel/kota opsional, budget opsional).
2. search_hotels — kalau user TIDAK sebut nama hotel/kota spesifik, panggil dengan query="" untuk list semua hotel.
3. list_rooms — untuk hotel kandidat (terutama yang termurah/sesuai preferensi).
4. check_availability — konfirmasi ketersediaan + dapatkan harga total.
5. create_reservation — langsung saja kalau user sudah jelas: ada tanggal + ada instruksi "termurah/apa saja/langsung booking".
6. pay_reservation — langsung setelah create_reservation sukses (gunakan reservation_id dari step 5).

Kapan boleh tanya user (HANYA kalau benar-benar ambigu):
- Tanggal tidak jelas atau tidak masuk akal.
- User tidak sebut durasi/check-out.
- Multiple hotel match dan user belum kasih kriteria untuk memilih (e.g. "termurah", "bintang 5", dst).

Kapan TIDAK perlu tanya:
- User bilang "apa saja", "yang termurah", "langsung booking", "tidak perlu konfirmasi" → langsung jalankan.
- Tanggal sudah jelas dan user minta "yang termurah" → pilih sendiri kamar termurah dari hasil list_rooms.

Pedoman:
- Selalu balas dalam bahasa Indonesia.
- Format tanggal: YYYY-MM-DD.
- Kalau check_availability gagal/penuh, otomatis coba kamar/tanggal alternatif sebelum tanya user.
- Setelah booking + pay sukses, tampilkan: reservation ID, total harga, dates, nama hotel, dan nomor kamar.`
)

// Orchestrator menjalankan agent loop untuk booking otomatis.
type Orchestrator struct {
	client    openai.Client
	model     string
	guestSvc  service.GuestService
	resSvc    service.ReservationService
	hotelRepo repository.HotelRepo
}

// NewOrchestrator membaca OPENAI_API_KEY dari .env / env var.
// Return error kalau API key tidak diset.
func NewOrchestrator(
	guestSvc service.GuestService,
	resSvc service.ReservationService,
	hotelRepo repository.HotelRepo,
) (*Orchestrator, error) {
	viper.SetConfigFile(".env")
	_ = viper.ReadInConfig()
	viper.AutomaticEnv()

	apiKey := viper.GetString("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY tidak diset (cek .env)")
	}

	model := viper.GetString("OPENAI_MODEL")
	if model == "" {
		model = defaultModel
	}

	return &Orchestrator{
		client:    openai.NewClient(option.WithAPIKey(apiKey)),
		model:     model,
		guestSvc:  guestSvc,
		resSvc:    resSvc,
		hotelRepo: hotelRepo,
	}, nil
}

// Result = output ke client.
type Result struct {
	Message       string           `json:"message"`                  // final reply AI ke user
	Iterations    int              `json:"iterations"`               // berapa kali round-trip ke OpenAI
	ToolCalls     []ToolCallRecord `json:"tool_calls"`               // trace lengkap apa yang AI lakukan
	ReservationID string           `json:"reservation_id,omitempty"` // jika ada booking sukses
	Model         string           `json:"model"`
}

// ToolCallRecord = jejak tiap pemanggilan tool, untuk transparency.
type ToolCallRecord struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
	Result    string         `json:"result"` // dipotong kalau terlalu panjang
	Error     string         `json:"error,omitempty"`
}

// RunBooking menjalankan agent loop sampai AI memberikan jawaban final
// atau mencapai max iterations.
func (o *Orchestrator) RunBooking(ctx context.Context, guestID, userMessage string) (*Result, error) {
	result := &Result{ToolCalls: []ToolCallRecord{}, Model: o.model}

	today := time.Now().Format("2006-01-02 (Monday)")
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(fmt.Sprintf(systemPrompt, today)),
		openai.UserMessage(userMessage),
	}
	tools := o.tools()

	for iter := 0; iter < maxIterations; iter++ {
		result.Iterations = iter + 1

		resp, err := o.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
			Model:    o.model,
			Messages: messages,
			Tools:    tools,
		})
		if err != nil {
			return nil, fmt.Errorf("openai call gagal: %w", err)
		}
		if len(resp.Choices) == 0 {
			return nil, fmt.Errorf("openai response tidak punya choices")
		}

		msg := resp.Choices[0].Message
		// Append assistant message ke history (penting untuk tool-call follow-up)
		messages = append(messages, msg.ToParam())

		// Tidak ada tool call → final answer
		if len(msg.ToolCalls) == 0 {
			result.Message = msg.Content
			return result, nil
		}

		// Eksekusi tiap tool call lalu append hasilnya
		for _, tc := range msg.ToolCalls {
			argMap := map[string]any{}
			_ = json.Unmarshal([]byte(tc.Function.Arguments), &argMap)

			toolResult, toolErr := o.executeTool(guestID, tc.Function.Name, argMap, result)

			rec := ToolCallRecord{
				Name:      tc.Function.Name,
				Arguments: argMap,
				Result:    truncate(toolResult, 500),
			}
			if toolErr != nil {
				rec.Error = toolErr.Error()
				toolResult = fmt.Sprintf(`{"error":%q}`, toolErr.Error())
			}
			result.ToolCalls = append(result.ToolCalls, rec)

			messages = append(messages, openai.ToolMessage(toolResult, tc.ID))
		}
	}

	return nil, fmt.Errorf("max iterations (%d) tercapai tanpa hasil final", maxIterations)
}

// ─── Tool dispatcher ─────────────────────────────────────────────────────────

func (o *Orchestrator) executeTool(guestID, name string, args map[string]any, result *Result) (string, error) {
	switch name {
	case "search_hotels":
		return o.searchHotels(args)
	case "list_rooms":
		return o.listRooms(args)
	case "check_availability":
		return o.checkAvailability(args)
	case "create_reservation":
		return o.createReservation(guestID, args, result)
	case "pay_reservation":
		return o.payReservation(guestID, args)
	default:
		return "", fmt.Errorf("tool tidak dikenali: %s", name)
	}
}

// ─── Tool implementations ────────────────────────────────────────────────────

func (o *Orchestrator) searchHotels(args map[string]any) (string, error) {
	query, _ := args["query"].(string)
	hotels, err := o.guestSvc.SearchHotels(query)
	if err != nil {
		return "", err
	}
	type item struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Address string `json:"address"`
		Status  string `json:"status"`
	}
	out := make([]item, 0, len(hotels))
	for _, h := range hotels {
		out = append(out, item{
			ID:      h.ID.String(),
			Name:    h.Name,
			Address: h.Address,
			Status:  h.Status,
		})
	}
	b, _ := json.Marshal(map[string]any{"hotels": out, "count": len(out)})
	return string(b), nil
}

func (o *Orchestrator) listRooms(args map[string]any) (string, error) {
	hotelID, _ := args["hotel_id"].(string)
	if hotelID == "" {
		return "", fmt.Errorf("hotel_id wajib")
	}
	rooms, err := o.hotelRepo.ListRooms(hotelID, "")
	if err != nil {
		return "", err
	}
	type item struct {
		ID            string  `json:"id"`
		RoomNumber    string  `json:"room_number"`
		RoomTypeID    string  `json:"room_type_id"`
		Status        string  `json:"status"`
		PricePerNight float64 `json:"price_per_night"`
		Capacity      int     `json:"capacity"`
		BedType       string  `json:"bed_type,omitempty"`
	}
	out := make([]item, 0, len(rooms))
	for _, r := range rooms {
		// Skip kamar tidak available langsung
		if r.Status != models.RoomStatusAvailable {
			continue
		}
		rt, err := o.hotelRepo.GetRoomTypeByID(r.RoomTypeID.String())
		if err != nil {
			continue
		}
		price := rt.PricePerNight
		if price == 0 {
			price = rt.BasePrice
		}
		out = append(out, item{
			ID:            r.ID.String(),
			RoomNumber:    r.RoomNumber,
			RoomTypeID:    r.RoomTypeID.String(),
			Status:        string(r.Status),
			PricePerNight: price,
			Capacity:      rt.Capacity,
			BedType:       string(rt.BedType),
		})
	}
	b, _ := json.Marshal(map[string]any{"rooms": out, "count": len(out)})
	return string(b), nil
}

func (o *Orchestrator) checkAvailability(args map[string]any) (string, error) {
	roomID, _ := args["room_id"].(string)
	checkInStr, _ := args["check_in"].(string)
	checkOutStr, _ := args["check_out"].(string)

	checkIn, err := time.Parse("2006-01-02", checkInStr)
	if err != nil {
		return "", fmt.Errorf("format check_in tidak valid (YYYY-MM-DD)")
	}
	checkOut, err := time.Parse("2006-01-02", checkOutStr)
	if err != nil {
		return "", fmt.Errorf("format check_out tidak valid (YYYY-MM-DD)")
	}

	quote, err := o.resSvc.QuoteReservation(roomID, checkIn, checkOut)
	if err != nil {
		return "", err
	}
	b, _ := json.Marshal(quote)
	return string(b), nil
}

func (o *Orchestrator) createReservation(guestID string, args map[string]any, result *Result) (string, error) {
	hotelID, _ := args["hotel_id"].(string)
	roomID, _ := args["room_id"].(string)
	checkInStr, _ := args["check_in"].(string)
	checkOutStr, _ := args["check_out"].(string)
	source, _ := args["booking_source"].(string)
	if source == "" {
		source = "online"
	}
	method, _ := args["payment_method"].(string)
	if method == "" {
		method = "transfer"
	}
	notes, _ := args["special_requests"].(string)

	checkIn, err := time.Parse("2006-01-02", checkInStr)
	if err != nil {
		return "", fmt.Errorf("format check_in tidak valid (YYYY-MM-DD)")
	}
	checkOut, err := time.Parse("2006-01-02", checkOutStr)
	if err != nil {
		return "", fmt.Errorf("format check_out tidak valid (YYYY-MM-DD)")
	}

	res, err := o.resSvc.CreateReservation(service.CreateReservationInput{
		GuestID:         guestID,
		HotelID:         hotelID,
		RoomID:          roomID,
		CheckIn:         checkIn,
		CheckOut:        checkOut,
		BookingSource:   source,
		PaymentMethod:   method,
		SpecialRequests: notes,
	})
	if err != nil {
		return "", err
	}

	// Simpan reservation_id untuk response final
	if res != nil && res.Reservation != nil {
		result.ReservationID = res.Reservation.ID.String()
	}
	b, _ := json.Marshal(res)
	return string(b), nil
}

func (o *Orchestrator) payReservation(guestID string, args map[string]any) (string, error) {
	reservationID, _ := args["reservation_id"].(string)
	method, _ := args["payment_method"].(string)
	if method == "" {
		method = "transfer"
	}
	res, tx, err := o.resSvc.MarkPaymentPaid(guestID, reservationID, method)
	if err != nil {
		return "", err
	}
	b, _ := json.Marshal(map[string]any{"reservation": res, "transaction": tx})
	return string(b), nil
}

// ─── Tool definitions (JSON schema) ──────────────────────────────────────────

func (o *Orchestrator) tools() []openai.ChatCompletionToolParam {
	return []openai.ChatCompletionToolParam{
		toolDef(
			"search_hotels",
			"Mencari hotel berdasarkan keyword (nama hotel atau kota). Return list hotel dengan id, name, address, city.",
			map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{
						"type":        "string",
						"description": "Keyword pencarian: nama hotel atau kota (contoh: 'Bali', 'Hotel Indonesia'). Boleh kosong untuk list semua.",
					},
				},
			},
		),
		toolDef(
			"list_rooms",
			"Mendapatkan daftar kamar yang tersedia (status=available) di sebuah hotel, lengkap dengan harga per malam dan kapasitas.",
			map[string]any{
				"type": "object",
				"properties": map[string]any{
					"hotel_id": map[string]any{
						"type":        "string",
						"description": "UUID hotel. Dapatkan dari search_hotels.",
					},
				},
				"required": []string{"hotel_id"},
			},
		),
		toolDef(
			"check_availability",
			"Mengecek apakah kamar tersedia di rentang tanggal tertentu, sekaligus dapatkan quote harga total.",
			map[string]any{
				"type": "object",
				"properties": map[string]any{
					"room_id":   map[string]any{"type": "string", "description": "UUID kamar"},
					"check_in":  map[string]any{"type": "string", "description": "Tanggal check-in YYYY-MM-DD"},
					"check_out": map[string]any{"type": "string", "description": "Tanggal check-out YYYY-MM-DD"},
				},
				"required": []string{"room_id", "check_in", "check_out"},
			},
		),
		toolDef(
			"create_reservation",
			"Membuat reservation BARU dengan status pending. PANGGIL HANYA SETELAH user konfirmasi. Return reservation_id yang dipakai untuk pay_reservation.",
			map[string]any{
				"type": "object",
				"properties": map[string]any{
					"hotel_id":         map[string]any{"type": "string", "description": "UUID hotel"},
					"room_id":          map[string]any{"type": "string", "description": "UUID kamar"},
					"check_in":         map[string]any{"type": "string", "description": "Tanggal check-in YYYY-MM-DD"},
					"check_out":        map[string]any{"type": "string", "description": "Tanggal check-out YYYY-MM-DD"},
					"booking_source":   map[string]any{"type": "string", "enum": []string{"online", "walk_in", "phone", "ota", "corporate"}, "description": "Default 'online'"},
					"payment_method":   map[string]any{"type": "string", "description": "Default 'transfer'"},
					"special_requests": map[string]any{"type": "string", "description": "Request khusus user (opsional)"},
				},
				"required": []string{"hotel_id", "room_id", "check_in", "check_out"},
			},
		),
		toolDef(
			"pay_reservation",
			"Memproses pembayaran reservation yang sudah dibuat. Panggil setelah create_reservation sukses.",
			map[string]any{
				"type": "object",
				"properties": map[string]any{
					"reservation_id": map[string]any{"type": "string", "description": "UUID reservation dari create_reservation"},
					"payment_method": map[string]any{"type": "string", "description": "Default 'transfer'"},
				},
				"required": []string{"reservation_id"},
			},
		),
	}
}

func toolDef(name, desc string, params map[string]any) openai.ChatCompletionToolParam {
	return openai.ChatCompletionToolParam{
		Function: shared.FunctionDefinitionParam{
			Name:        name,
			Description: openai.String(desc),
			Parameters:  shared.FunctionParameters(params),
		},
	}
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "...(truncated)"
}

// Validate konfigurasi awal — bisa dipanggil di main.go untuk fail-fast.
func ValidateConfig() error {
	viper.SetConfigFile(".env")
	_ = viper.ReadInConfig()
	viper.AutomaticEnv()
	if strings.TrimSpace(viper.GetString("OPENAI_API_KEY")) == "" {
		return fmt.Errorf("OPENAI_API_KEY tidak diset")
	}
	return nil
}
