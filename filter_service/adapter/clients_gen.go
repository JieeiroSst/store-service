package adapter

import (
	"net/http"
	"os"

	"github.com/JIeeiroSst/filter-service/adapter/account_transaction_service"
	"github.com/JIeeiroSst/filter-service/adapter/address_country_service"
	"github.com/JIeeiroSst/filter-service/adapter/admanagement_service"
	"github.com/JIeeiroSst/filter-service/adapter/ai_agent_system"
	"github.com/JIeeiroSst/filter-service/adapter/airflow_service"
	"github.com/JIeeiroSst/filter-service/adapter/arrange_service"
	"github.com/JIeeiroSst/filter-service/adapter/authorize_service"
	"github.com/JIeeiroSst/filter-service/adapter/automatic_payment_service"
	"github.com/JIeeiroSst/filter-service/adapter/banking_service"
	"github.com/JIeeiroSst/filter-service/adapter/basket_service"
	"github.com/JIeeiroSst/filter-service/adapter/billing_service"
	"github.com/JIeeiroSst/filter-service/adapter/bonuslink_service"
	"github.com/JIeeiroSst/filter-service/adapter/book_service"
	"github.com/JIeeiroSst/filter-service/adapter/book_store_service"
	"github.com/JIeeiroSst/filter-service/adapter/booking_mini_service"
	"github.com/JIeeiroSst/filter-service/adapter/bot_service"
	"github.com/JIeeiroSst/filter-service/adapter/calculate_service"
	"github.com/JIeeiroSst/filter-service/adapter/call_center_ai"
	"github.com/JIeeiroSst/filter-service/adapter/car_rental_service"
	"github.com/JIeeiroSst/filter-service/adapter/catalogues_service"
	"github.com/JIeeiroSst/filter-service/adapter/cdn_service"
	"github.com/JIeeiroSst/filter-service/adapter/chat_service"
	"github.com/JIeeiroSst/filter-service/adapter/chatbot_system"
	"github.com/JIeeiroSst/filter-service/adapter/coupon_service"
	"github.com/JIeeiroSst/filter-service/adapter/customer_relationship_service"
	"github.com/JIeeiroSst/filter-service/adapter/doordash_service"
	"github.com/JIeeiroSst/filter-service/adapter/draw_image_service"
	"github.com/JIeeiroSst/filter-service/adapter/ekyc_service"
	"github.com/JIeeiroSst/filter-service/adapter/face_recognition_service"
	"github.com/JIeeiroSst/filter-service/adapter/food_service"
	"github.com/JIeeiroSst/filter-service/adapter/geoservice"
	"github.com/JIeeiroSst/filter-service/adapter/hospital_patient_management_service"
	"github.com/JIeeiroSst/filter-service/adapter/identified_service"
	"github.com/JIeeiroSst/filter-service/adapter/integrated_payment_service"
	"github.com/JIeeiroSst/filter-service/adapter/kms_service"
	"github.com/JIeeiroSst/filter-service/adapter/livestream_service"
	"github.com/JIeeiroSst/filter-service/adapter/lottery_service"
	"github.com/JIeeiroSst/filter-service/adapter/manage_service"
	"github.com/JIeeiroSst/filter-service/adapter/media_service"
	"github.com/JIeeiroSst/filter-service/adapter/medical_service"
	"github.com/JIeeiroSst/filter-service/adapter/movie_recommendation_service"
	"github.com/JIeeiroSst/filter-service/adapter/notification_service"
	"github.com/JIeeiroSst/filter-service/adapter/notifyhub_service"
	"github.com/JIeeiroSst/filter-service/adapter/ollama_model_py"
	"github.com/JIeeiroSst/filter-service/adapter/ollama_service"
	"github.com/JIeeiroSst/filter-service/adapter/parking_lot_service"
	"github.com/JIeeiroSst/filter-service/adapter/partner_service"
	"github.com/JIeeiroSst/filter-service/adapter/payer_service"
	"github.com/JIeeiroSst/filter-service/adapter/payment_service"
	"github.com/JIeeiroSst/filter-service/adapter/payment_wallet_service"
	"github.com/JIeeiroSst/filter-service/adapter/photo_service"
	"github.com/JIeeiroSst/filter-service/adapter/point_service"
	"github.com/JIeeiroSst/filter-service/adapter/polymarket_service"
	"github.com/JIeeiroSst/filter-service/adapter/post_service"
	"github.com/JIeeiroSst/filter-service/adapter/qr_service"
	"github.com/JIeeiroSst/filter-service/adapter/recompense_service"
	"github.com/JIeeiroSst/filter-service/adapter/recruitment_platform_service"
	"github.com/JIeeiroSst/filter-service/adapter/referral_service"
	"github.com/JIeeiroSst/filter-service/adapter/rent_house_service"
	"github.com/JIeeiroSst/filter-service/adapter/reservation_service"
	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/adapter/restaurant_accounting_service"
	"github.com/JIeeiroSst/filter-service/adapter/restaurant_consumer_service"
	"github.com/JIeeiroSst/filter-service/adapter/restaurant_delivery_service"
	"github.com/JIeeiroSst/filter-service/adapter/restaurant_kitchen_service"
	"github.com/JIeeiroSst/filter-service/adapter/restaurant_order_service"
	"github.com/JIeeiroSst/filter-service/adapter/reward_service"
	"github.com/JIeeiroSst/filter-service/adapter/room_service"
	"github.com/JIeeiroSst/filter-service/adapter/sard_service"
	"github.com/JIeeiroSst/filter-service/adapter/shipping_service"
	"github.com/JIeeiroSst/filter-service/adapter/shopify_service"
	"github.com/JIeeiroSst/filter-service/adapter/shortlink_service"
	"github.com/JIeeiroSst/filter-service/adapter/solana_service"
	"github.com/JIeeiroSst/filter-service/adapter/threads_service"
	"github.com/JIeeiroSst/filter-service/adapter/ticket_service"
	"github.com/JIeeiroSst/filter-service/adapter/toggle_service"
	"github.com/JIeeiroSst/filter-service/adapter/tool_service"
	"github.com/JIeeiroSst/filter-service/adapter/upload_service"
	"github.com/JIeeiroSst/filter-service/adapter/user_service"
	"github.com/JIeeiroSst/filter-service/adapter/vending_machine_service"
	"github.com/JIeeiroSst/filter-service/adapter/video_service"
	"github.com/JIeeiroSst/filter-service/adapter/voucher_service"
	"github.com/JIeeiroSst/filter-service/adapter/wallet_service"
	"github.com/JIeeiroSst/filter-service/adapter/webrtc_service"
	"github.com/JIeeiroSst/filter-service/adapter/wishlists_service"
)

type ServiceInfo struct {
	Name      string
	Field     string
	EnvVar    string
	BaseURL   string
	Endpoints int
}

type Clients struct {
	Services                         []ServiceInfo
	AccountTransactionService        *account_transaction_service.Client
	AddressCountryService            *address_country_service.Client
	AdmanagementService              *admanagement_service.Client
	AiAgentSystem                    *ai_agent_system.Client
	AirflowService                   *airflow_service.Client
	ArrangeService                   *arrange_service.Client
	AuthorizeService                 *authorize_service.Client
	AutomaticPaymentService          *automatic_payment_service.Client
	BankingService                   *banking_service.Client
	BasketService                    *basket_service.Client
	BillingService                   *billing_service.Client
	BonuslinkService                 *bonuslink_service.Client
	BookService                      *book_service.Client
	BookStoreService                 *book_store_service.Client
	BookingMiniService               *booking_mini_service.Client
	BotService                       *bot_service.Client
	CalculateService                 *calculate_service.Client
	CallCenterAi                     *call_center_ai.Client
	CarRentalService                 *car_rental_service.Client
	CataloguesService                *catalogues_service.Client
	CdnService                       *cdn_service.Client
	ChatService                      *chat_service.Client
	ChatbotSystem                    *chatbot_system.Client
	CouponService                    *coupon_service.Client
	CustomerRelationshipService      *customer_relationship_service.Client
	DoordashService                  *doordash_service.Client
	DrawImageService                 *draw_image_service.Client
	EkycService                      *ekyc_service.Client
	FaceRecognitionService           *face_recognition_service.Client
	FoodService                      *food_service.Client
	Geoservice                       *geoservice.Client
	HospitalPatientManagementService *hospital_patient_management_service.Client
	IdentifiedService                *identified_service.Client
	IntegratedPaymentService         *integrated_payment_service.Client
	KmsService                       *kms_service.Client
	LivestreamService                *livestream_service.Client
	LotteryService                   *lottery_service.Client
	ManageService                    *manage_service.Client
	MediaService                     *media_service.Client
	MedicalService                   *medical_service.Client
	MovieRecommendationService       *movie_recommendation_service.Client
	NotificationService              *notification_service.Client
	NotifyhubService                 *notifyhub_service.Client
	OllamaModelPy                    *ollama_model_py.Client
	OllamaService                    *ollama_service.Client
	ParkingLotService                *parking_lot_service.Client
	PartnerService                   *partner_service.Client
	PayerService                     *payer_service.Client
	PaymentWalletService             *payment_wallet_service.Client
	PaymentService                   *payment_service.Client
	PhotoService                     *photo_service.Client
	PointService                     *point_service.Client
	PolymarketService                *polymarket_service.Client
	PostService                      *post_service.Client
	QRService                        *qr_service.Client
	RecompenseService                *recompense_service.Client
	RecruitmentPlatformService       *recruitment_platform_service.Client
	ReferralService                  *referral_service.Client
	RentHouseService                 *rent_house_service.Client
	ReservationService               *reservation_service.Client
	RestaurantAccountingService      *restaurant_accounting_service.Client
	RestaurantConsumerService        *restaurant_consumer_service.Client
	RestaurantDeliveryService        *restaurant_delivery_service.Client
	RestaurantKitchenService         *restaurant_kitchen_service.Client
	RestaurantOrderService           *restaurant_order_service.Client
	RewardService                    *reward_service.Client
	RoomService                      *room_service.Client
	SardService                      *sard_service.Client
	ShippingService                  *shipping_service.Client
	ShopifyService                   *shopify_service.Client
	ShortlinkService                 *shortlink_service.Client
	SolanaService                    *solana_service.Client
	ThreadsService                   *threads_service.Client
	TicketService                    *ticket_service.Client
	ToggleService                    *toggle_service.Client
	ToolService                      *tool_service.Client
	UploadService                    *upload_service.Client
	UserService                      *user_service.Client
	VendingMachineService            *vending_machine_service.Client
	VideoService                     *video_service.Client
	VoucherService                   *voucher_service.Client
	WalletService                    *wallet_service.Client
	WebrtcService                    *webrtc_service.Client
	WishlistsService                 *wishlists_service.Client
}

func baseURL(env, fallback string) string {
	if v := os.Getenv(env); v != "" {
		return v
	}
	return fallback
}

func NewClients(httpClient *http.Client) *Clients {
	c := &Clients{}
	{
		u := baseURL("ACCOUNT_TRANSACTION_SERVICE_URL", account_transaction_service.DefaultBaseURL)
		c.AccountTransactionService = account_transaction_service.New(rest.New(account_transaction_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: account_transaction_service.Service, Field: "accountTransactionService", EnvVar: "ACCOUNT_TRANSACTION_SERVICE_URL", BaseURL: u, Endpoints: 5})
	}
	{
		u := baseURL("ADDRESS_COUNTRY_SERVICE_URL", address_country_service.DefaultBaseURL)
		c.AddressCountryService = address_country_service.New(rest.New(address_country_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: address_country_service.Service, Field: "addressCountryService", EnvVar: "ADDRESS_COUNTRY_SERVICE_URL", BaseURL: u, Endpoints: 14})
	}
	{
		u := baseURL("ADMANAGEMENT_SERVICE_URL", admanagement_service.DefaultBaseURL)
		c.AdmanagementService = admanagement_service.New(rest.New(admanagement_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: admanagement_service.Service, Field: "admanagementService", EnvVar: "ADMANAGEMENT_SERVICE_URL", BaseURL: u, Endpoints: 23})
	}
	{
		u := baseURL("AI_AGENT_SYSTEM_URL", ai_agent_system.DefaultBaseURL)
		c.AiAgentSystem = ai_agent_system.New(rest.New(ai_agent_system.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: ai_agent_system.Service, Field: "aiAgentSystem", EnvVar: "AI_AGENT_SYSTEM_URL", BaseURL: u, Endpoints: 2})
	}
	{
		u := baseURL("AIRFLOW_SERVICE_URL", airflow_service.DefaultBaseURL)
		c.AirflowService = airflow_service.New(rest.New(airflow_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: airflow_service.Service, Field: "airflowService", EnvVar: "AIRFLOW_SERVICE_URL", BaseURL: u, Endpoints: 11})
	}
	{
		u := baseURL("ARRANGE_SERVICE_URL", arrange_service.DefaultBaseURL)
		c.ArrangeService = arrange_service.New(rest.New(arrange_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: arrange_service.Service, Field: "arrangeService", EnvVar: "ARRANGE_SERVICE_URL", BaseURL: u, Endpoints: 4})
	}
	{
		u := baseURL("AUTHORIZE_SERVICE_URL", authorize_service.DefaultBaseURL)
		c.AuthorizeService = authorize_service.New(rest.New(authorize_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: authorize_service.Service, Field: "authorizeService", EnvVar: "AUTHORIZE_SERVICE_URL", BaseURL: u, Endpoints: 2})
	}
	{
		u := baseURL("AUTOMATIC_PAYMENT_SERVICE_URL", automatic_payment_service.DefaultBaseURL)
		c.AutomaticPaymentService = automatic_payment_service.New(rest.New(automatic_payment_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: automatic_payment_service.Service, Field: "automaticPaymentService", EnvVar: "AUTOMATIC_PAYMENT_SERVICE_URL", BaseURL: u, Endpoints: 6})
	}
	{
		u := baseURL("BANKING_SERVICE_URL", banking_service.DefaultBaseURL)
		c.BankingService = banking_service.New(rest.New(banking_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: banking_service.Service, Field: "bankingService", EnvVar: "BANKING_SERVICE_URL", BaseURL: u, Endpoints: 16})
	}
	{
		u := baseURL("BASKET_SERVICE_URL", basket_service.DefaultBaseURL)
		c.BasketService = basket_service.New(rest.New(basket_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: basket_service.Service, Field: "basketService", EnvVar: "BASKET_SERVICE_URL", BaseURL: u, Endpoints: 9})
	}
	{
		u := baseURL("BILLING_SERVICE_URL", billing_service.DefaultBaseURL)
		c.BillingService = billing_service.New(rest.New(billing_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: billing_service.Service, Field: "billingService", EnvVar: "BILLING_SERVICE_URL", BaseURL: u, Endpoints: 12})
	}
	{
		u := baseURL("BONUSLINK_SERVICE_URL", bonuslink_service.DefaultBaseURL)
		c.BonuslinkService = bonuslink_service.New(rest.New(bonuslink_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: bonuslink_service.Service, Field: "bonuslinkService", EnvVar: "BONUSLINK_SERVICE_URL", BaseURL: u, Endpoints: 2})
	}
	{
		u := baseURL("BOOK_SERVICE_URL", book_service.DefaultBaseURL)
		c.BookService = book_service.New(rest.New(book_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: book_service.Service, Field: "bookService", EnvVar: "BOOK_SERVICE_URL", BaseURL: u, Endpoints: 7})
	}
	{
		u := baseURL("BOOK_STORE_SERVICE_URL", book_store_service.DefaultBaseURL)
		c.BookStoreService = book_store_service.New(rest.New(book_store_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: book_store_service.Service, Field: "bookStoreService", EnvVar: "BOOK_STORE_SERVICE_URL", BaseURL: u, Endpoints: 13})
	}
	{
		u := baseURL("BOOKING_MINI_SERVICE_URL", booking_mini_service.DefaultBaseURL)
		c.BookingMiniService = booking_mini_service.New(rest.New(booking_mini_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: booking_mini_service.Service, Field: "bookingMiniService", EnvVar: "BOOKING_MINI_SERVICE_URL", BaseURL: u, Endpoints: 3})
	}
	{
		u := baseURL("BOT_SERVICE_URL", bot_service.DefaultBaseURL)
		c.BotService = bot_service.New(rest.New(bot_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: bot_service.Service, Field: "botService", EnvVar: "BOT_SERVICE_URL", BaseURL: u, Endpoints: 2})
	}
	{
		u := baseURL("CALCULATE_SERVICE_URL", calculate_service.DefaultBaseURL)
		c.CalculateService = calculate_service.New(rest.New(calculate_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: calculate_service.Service, Field: "calculateService", EnvVar: "CALCULATE_SERVICE_URL", BaseURL: u, Endpoints: 6})
	}
	{
		u := baseURL("CALL_CENTER_AI_URL", call_center_ai.DefaultBaseURL)
		c.CallCenterAi = call_center_ai.New(rest.New(call_center_ai.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: call_center_ai.Service, Field: "callCenterAi", EnvVar: "CALL_CENTER_AI_URL", BaseURL: u, Endpoints: 6})
	}
	{
		u := baseURL("CAR_RENTAL_SERVICE_URL", car_rental_service.DefaultBaseURL)
		c.CarRentalService = car_rental_service.New(rest.New(car_rental_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: car_rental_service.Service, Field: "carRentalService", EnvVar: "CAR_RENTAL_SERVICE_URL", BaseURL: u, Endpoints: 10})
	}
	{
		u := baseURL("CATALOGUES_SERVICE_URL", catalogues_service.DefaultBaseURL)
		c.CataloguesService = catalogues_service.New(rest.New(catalogues_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: catalogues_service.Service, Field: "cataloguesService", EnvVar: "CATALOGUES_SERVICE_URL", BaseURL: u, Endpoints: 8})
	}
	{
		u := baseURL("CDN_SERVICE_URL", cdn_service.DefaultBaseURL)
		c.CdnService = cdn_service.New(rest.New(cdn_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: cdn_service.Service, Field: "cdnService", EnvVar: "CDN_SERVICE_URL", BaseURL: u, Endpoints: 2})
	}
	{
		u := baseURL("CHAT_SERVICE_URL", chat_service.DefaultBaseURL)
		c.ChatService = chat_service.New(rest.New(chat_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: chat_service.Service, Field: "chatService", EnvVar: "CHAT_SERVICE_URL", BaseURL: u, Endpoints: 2})
	}
	{
		u := baseURL("CHATBOT_SYSTEM_URL", chatbot_system.DefaultBaseURL)
		c.ChatbotSystem = chatbot_system.New(rest.New(chatbot_system.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: chatbot_system.Service, Field: "chatbotSystem", EnvVar: "CHATBOT_SYSTEM_URL", BaseURL: u, Endpoints: 2})
	}
	{
		u := baseURL("COUPON_SERVICE_URL", coupon_service.DefaultBaseURL)
		c.CouponService = coupon_service.New(rest.New(coupon_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: coupon_service.Service, Field: "couponService", EnvVar: "COUPON_SERVICE_URL", BaseURL: u, Endpoints: 8})
	}
	{
		u := baseURL("CUSTOMER_RELATIONSHIP_SERVICE_URL", customer_relationship_service.DefaultBaseURL)
		c.CustomerRelationshipService = customer_relationship_service.New(rest.New(customer_relationship_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: customer_relationship_service.Service, Field: "customerRelationshipService", EnvVar: "CUSTOMER_RELATIONSHIP_SERVICE_URL", BaseURL: u, Endpoints: 28})
	}
	{
		u := baseURL("DOORDASH_SERVICE_URL", doordash_service.DefaultBaseURL)
		c.DoordashService = doordash_service.New(rest.New(doordash_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: doordash_service.Service, Field: "doordashService", EnvVar: "DOORDASH_SERVICE_URL", BaseURL: u, Endpoints: 4})
	}
	{
		u := baseURL("DRAW_IMAGE_SERVICE_URL", draw_image_service.DefaultBaseURL)
		c.DrawImageService = draw_image_service.New(rest.New(draw_image_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: draw_image_service.Service, Field: "drawImageService", EnvVar: "DRAW_IMAGE_SERVICE_URL", BaseURL: u, Endpoints: 1})
	}
	{
		u := baseURL("EKYC_SERVICE_URL", ekyc_service.DefaultBaseURL)
		c.EkycService = ekyc_service.New(rest.New(ekyc_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: ekyc_service.Service, Field: "ekycService", EnvVar: "EKYC_SERVICE_URL", BaseURL: u, Endpoints: 1})
	}
	{
		u := baseURL("FACE_RECOGNITION_SERVICE_URL", face_recognition_service.DefaultBaseURL)
		c.FaceRecognitionService = face_recognition_service.New(rest.New(face_recognition_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: face_recognition_service.Service, Field: "faceRecognitionService", EnvVar: "FACE_RECOGNITION_SERVICE_URL", BaseURL: u, Endpoints: 1})
	}
	{
		u := baseURL("FOOD_SERVICE_URL", food_service.DefaultBaseURL)
		c.FoodService = food_service.New(rest.New(food_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: food_service.Service, Field: "foodService", EnvVar: "FOOD_SERVICE_URL", BaseURL: u, Endpoints: 3})
	}
	{
		u := baseURL("GEOSERVICE_URL", geoservice.DefaultBaseURL)
		c.Geoservice = geoservice.New(rest.New(geoservice.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: geoservice.Service, Field: "geoservice", EnvVar: "GEOSERVICE_URL", BaseURL: u, Endpoints: 9})
	}
	{
		u := baseURL("HOSPITAL_PATIENT_MANAGEMENT_SERVICE_URL", hospital_patient_management_service.DefaultBaseURL)
		c.HospitalPatientManagementService = hospital_patient_management_service.New(rest.New(hospital_patient_management_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: hospital_patient_management_service.Service, Field: "hospitalPatientManagementService", EnvVar: "HOSPITAL_PATIENT_MANAGEMENT_SERVICE_URL", BaseURL: u, Endpoints: 21})
	}
	{
		u := baseURL("IDENTIFIED_SERVICE_URL", identified_service.DefaultBaseURL)
		c.IdentifiedService = identified_service.New(rest.New(identified_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: identified_service.Service, Field: "identifiedService", EnvVar: "IDENTIFIED_SERVICE_URL", BaseURL: u, Endpoints: 3})
	}
	{
		u := baseURL("INTEGRATED_PAYMENT_SERVICE_URL", integrated_payment_service.DefaultBaseURL)
		c.IntegratedPaymentService = integrated_payment_service.New(rest.New(integrated_payment_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: integrated_payment_service.Service, Field: "integratedPaymentService", EnvVar: "INTEGRATED_PAYMENT_SERVICE_URL", BaseURL: u, Endpoints: 2})
	}
	{
		u := baseURL("KMS_SERVICE_URL", kms_service.DefaultBaseURL)
		c.KmsService = kms_service.New(rest.New(kms_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: kms_service.Service, Field: "kmsService", EnvVar: "KMS_SERVICE_URL", BaseURL: u, Endpoints: 6})
	}
	{
		u := baseURL("LIVESTREAM_SERVICE_URL", livestream_service.DefaultBaseURL)
		c.LivestreamService = livestream_service.New(rest.New(livestream_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: livestream_service.Service, Field: "livestreamService", EnvVar: "LIVESTREAM_SERVICE_URL", BaseURL: u, Endpoints: 6})
	}
	{
		u := baseURL("LOTTERY_SERVICE_URL", lottery_service.DefaultBaseURL)
		c.LotteryService = lottery_service.New(rest.New(lottery_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: lottery_service.Service, Field: "lotteryService", EnvVar: "LOTTERY_SERVICE_URL", BaseURL: u, Endpoints: 1})
	}
	{
		u := baseURL("MANAGE_SERVICE_URL", manage_service.DefaultBaseURL)
		c.ManageService = manage_service.New(rest.New(manage_service.Service, u, httpClient, "X-Api-Key"))
		c.Services = append(c.Services, ServiceInfo{Name: manage_service.Service, Field: "manageService", EnvVar: "MANAGE_SERVICE_URL", BaseURL: u, Endpoints: 89})
	}
	{
		u := baseURL("MEDIA_SERVICE_URL", media_service.DefaultBaseURL)
		c.MediaService = media_service.New(rest.New(media_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: media_service.Service, Field: "mediaService", EnvVar: "MEDIA_SERVICE_URL", BaseURL: u, Endpoints: 1})
	}
	{
		u := baseURL("MEDICAL_SERVICE_URL", medical_service.DefaultBaseURL)
		c.MedicalService = medical_service.New(rest.New(medical_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: medical_service.Service, Field: "medicalService", EnvVar: "MEDICAL_SERVICE_URL", BaseURL: u, Endpoints: 10})
	}
	{
		u := baseURL("MOVIE_RECOMMENDATION_SERVICE_URL", movie_recommendation_service.DefaultBaseURL)
		c.MovieRecommendationService = movie_recommendation_service.New(rest.New(movie_recommendation_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: movie_recommendation_service.Service, Field: "movieRecommendationService", EnvVar: "MOVIE_RECOMMENDATION_SERVICE_URL", BaseURL: u, Endpoints: 7})
	}
	{
		u := baseURL("NOTIFICATION_SERVICE_URL", notification_service.DefaultBaseURL)
		c.NotificationService = notification_service.New(rest.New(notification_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: notification_service.Service, Field: "notificationService", EnvVar: "NOTIFICATION_SERVICE_URL", BaseURL: u, Endpoints: 9})
	}
	{
		u := baseURL("NOTIFYHUB_SERVICE_URL", notifyhub_service.DefaultBaseURL)
		c.NotifyhubService = notifyhub_service.New(rest.New(notifyhub_service.Service, u, httpClient, "X-API-Key"))
		c.Services = append(c.Services, ServiceInfo{Name: notifyhub_service.Service, Field: "notifyhubService", EnvVar: "NOTIFYHUB_SERVICE_URL", BaseURL: u, Endpoints: 10})
	}
	{
		u := baseURL("OLLAMA_MODEL_PY_URL", ollama_model_py.DefaultBaseURL)
		c.OllamaModelPy = ollama_model_py.New(rest.New(ollama_model_py.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: ollama_model_py.Service, Field: "ollamaModelPy", EnvVar: "OLLAMA_MODEL_PY_URL", BaseURL: u, Endpoints: 1})
	}
	{
		u := baseURL("OLLAMA_SERVICE_URL", ollama_service.DefaultBaseURL)
		c.OllamaService = ollama_service.New(rest.New(ollama_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: ollama_service.Service, Field: "ollamaService", EnvVar: "OLLAMA_SERVICE_URL", BaseURL: u, Endpoints: 5})
	}
	{
		u := baseURL("PARKING_LOT_SERVICE_URL", parking_lot_service.DefaultBaseURL)
		c.ParkingLotService = parking_lot_service.New(rest.New(parking_lot_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: parking_lot_service.Service, Field: "parkingLotService", EnvVar: "PARKING_LOT_SERVICE_URL", BaseURL: u, Endpoints: 4})
	}
	{
		u := baseURL("PARTNER_SERVICE_URL", partner_service.DefaultBaseURL)
		c.PartnerService = partner_service.New(rest.New(partner_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: partner_service.Service, Field: "partnerService", EnvVar: "PARTNER_SERVICE_URL", BaseURL: u, Endpoints: 11})
	}
	{
		u := baseURL("PAYER_SERVICE_URL", payer_service.DefaultBaseURL)
		c.PayerService = payer_service.New(rest.New(payer_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: payer_service.Service, Field: "payerService", EnvVar: "PAYER_SERVICE_URL", BaseURL: u, Endpoints: 1})
	}
	{
		u := baseURL("PAYMENT_WALLET_SERVICE_URL", payment_wallet_service.DefaultBaseURL)
		c.PaymentWalletService = payment_wallet_service.New(rest.New(payment_wallet_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: payment_wallet_service.Service, Field: "paymentWalletService", EnvVar: "PAYMENT_WALLET_SERVICE_URL", BaseURL: u, Endpoints: 11})
	}
	{
		u := baseURL("PAYMENT_SERVICE_URL", payment_service.DefaultBaseURL)
		c.PaymentService = payment_service.New(rest.New(payment_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: payment_service.Service, Field: "paymentService", EnvVar: "PAYMENT_SERVICE_URL", BaseURL: u, Endpoints: 3})
	}
	{
		u := baseURL("PHOTO_SERVICE_URL", photo_service.DefaultBaseURL)
		c.PhotoService = photo_service.New(rest.New(photo_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: photo_service.Service, Field: "photoService", EnvVar: "PHOTO_SERVICE_URL", BaseURL: u, Endpoints: 1})
	}
	{
		u := baseURL("POINT_SERVICE_URL", point_service.DefaultBaseURL)
		c.PointService = point_service.New(rest.New(point_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: point_service.Service, Field: "pointService", EnvVar: "POINT_SERVICE_URL", BaseURL: u, Endpoints: 6})
	}
	{
		u := baseURL("POLYMARKET_SERVICE_URL", polymarket_service.DefaultBaseURL)
		c.PolymarketService = polymarket_service.New(rest.New(polymarket_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: polymarket_service.Service, Field: "polymarketService", EnvVar: "POLYMARKET_SERVICE_URL", BaseURL: u, Endpoints: 26})
	}
	{
		u := baseURL("POST_SERVICE_URL", post_service.DefaultBaseURL)
		c.PostService = post_service.New(rest.New(post_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: post_service.Service, Field: "postService", EnvVar: "POST_SERVICE_URL", BaseURL: u, Endpoints: 4})
	}
	{
		u := baseURL("QR_SERVICE_URL", qr_service.DefaultBaseURL)
		c.QRService = qr_service.New(rest.New(qr_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: qr_service.Service, Field: "qrService", EnvVar: "QR_SERVICE_URL", BaseURL: u, Endpoints: 4})
	}
	{
		u := baseURL("RECOMPENSE_SERVICE_URL", recompense_service.DefaultBaseURL)
		c.RecompenseService = recompense_service.New(rest.New(recompense_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: recompense_service.Service, Field: "recompenseService", EnvVar: "RECOMPENSE_SERVICE_URL", BaseURL: u, Endpoints: 5})
	}
	{
		u := baseURL("RECRUITMENT_PLATFORM_SERVICE_URL", recruitment_platform_service.DefaultBaseURL)
		c.RecruitmentPlatformService = recruitment_platform_service.New(rest.New(recruitment_platform_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: recruitment_platform_service.Service, Field: "recruitmentPlatformService", EnvVar: "RECRUITMENT_PLATFORM_SERVICE_URL", BaseURL: u, Endpoints: 16})
	}
	{
		u := baseURL("REFERRAL_SERVICE_URL", referral_service.DefaultBaseURL)
		c.ReferralService = referral_service.New(rest.New(referral_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: referral_service.Service, Field: "referralService", EnvVar: "REFERRAL_SERVICE_URL", BaseURL: u, Endpoints: 6})
	}
	{
		u := baseURL("RENT_HOUSE_SERVICE_URL", rent_house_service.DefaultBaseURL)
		c.RentHouseService = rent_house_service.New(rest.New(rent_house_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: rent_house_service.Service, Field: "rentHouseService", EnvVar: "RENT_HOUSE_SERVICE_URL", BaseURL: u, Endpoints: 17})
	}
	{
		u := baseURL("RESERVATION_SERVICE_URL", reservation_service.DefaultBaseURL)
		c.ReservationService = reservation_service.New(rest.New(reservation_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: reservation_service.Service, Field: "reservationService", EnvVar: "RESERVATION_SERVICE_URL", BaseURL: u, Endpoints: 23})
	}
	{
		u := baseURL("RESTAURANT_ACCOUNTING_SERVICE_URL", restaurant_accounting_service.DefaultBaseURL)
		c.RestaurantAccountingService = restaurant_accounting_service.New(rest.New(restaurant_accounting_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: restaurant_accounting_service.Service, Field: "restaurantAccountingService", EnvVar: "RESTAURANT_ACCOUNTING_SERVICE_URL", BaseURL: u, Endpoints: 1})
	}
	{
		u := baseURL("RESTAURANT_CONSUMER_SERVICE_URL", restaurant_consumer_service.DefaultBaseURL)
		c.RestaurantConsumerService = restaurant_consumer_service.New(rest.New(restaurant_consumer_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: restaurant_consumer_service.Service, Field: "restaurantConsumerService", EnvVar: "RESTAURANT_CONSUMER_SERVICE_URL", BaseURL: u, Endpoints: 1})
	}
	{
		u := baseURL("RESTAURANT_DELIVERY_SERVICE_URL", restaurant_delivery_service.DefaultBaseURL)
		c.RestaurantDeliveryService = restaurant_delivery_service.New(rest.New(restaurant_delivery_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: restaurant_delivery_service.Service, Field: "restaurantDeliveryService", EnvVar: "RESTAURANT_DELIVERY_SERVICE_URL", BaseURL: u, Endpoints: 1})
	}
	{
		u := baseURL("RESTAURANT_KITCHEN_SERVICE_URL", restaurant_kitchen_service.DefaultBaseURL)
		c.RestaurantKitchenService = restaurant_kitchen_service.New(rest.New(restaurant_kitchen_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: restaurant_kitchen_service.Service, Field: "restaurantKitchenService", EnvVar: "RESTAURANT_KITCHEN_SERVICE_URL", BaseURL: u, Endpoints: 4})
	}
	{
		u := baseURL("RESTAURANT_ORDER_SERVICE_URL", restaurant_order_service.DefaultBaseURL)
		c.RestaurantOrderService = restaurant_order_service.New(rest.New(restaurant_order_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: restaurant_order_service.Service, Field: "restaurantOrderService", EnvVar: "RESTAURANT_ORDER_SERVICE_URL", BaseURL: u, Endpoints: 4})
	}
	{
		u := baseURL("REWARD_SERVICE_URL", reward_service.DefaultBaseURL)
		c.RewardService = reward_service.New(rest.New(reward_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: reward_service.Service, Field: "rewardService", EnvVar: "REWARD_SERVICE_URL", BaseURL: u, Endpoints: 2})
	}
	{
		u := baseURL("ROOM_SERVICE_URL", room_service.DefaultBaseURL)
		c.RoomService = room_service.New(rest.New(room_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: room_service.Service, Field: "roomService", EnvVar: "ROOM_SERVICE_URL", BaseURL: u, Endpoints: 5})
	}
	{
		u := baseURL("SARD_SERVICE_URL", sard_service.DefaultBaseURL)
		c.SardService = sard_service.New(rest.New(sard_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: sard_service.Service, Field: "sardService", EnvVar: "SARD_SERVICE_URL", BaseURL: u, Endpoints: 1})
	}
	{
		u := baseURL("SHIPPING_SERVICE_URL", shipping_service.DefaultBaseURL)
		c.ShippingService = shipping_service.New(rest.New(shipping_service.Service, u, httpClient, "X-API-Key"))
		c.Services = append(c.Services, ServiceInfo{Name: shipping_service.Service, Field: "shippingService", EnvVar: "SHIPPING_SERVICE_URL", BaseURL: u, Endpoints: 7})
	}
	{
		u := baseURL("SHOPIFY_SERVICE_URL", shopify_service.DefaultBaseURL)
		c.ShopifyService = shopify_service.New(rest.New(shopify_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: shopify_service.Service, Field: "shopifyService", EnvVar: "SHOPIFY_SERVICE_URL", BaseURL: u, Endpoints: 4})
	}
	{
		u := baseURL("SHORTLINK_SERVICE_URL", shortlink_service.DefaultBaseURL)
		c.ShortlinkService = shortlink_service.New(rest.New(shortlink_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: shortlink_service.Service, Field: "shortlinkService", EnvVar: "SHORTLINK_SERVICE_URL", BaseURL: u, Endpoints: 14})
	}
	{
		u := baseURL("SOLANA_SERVICE_URL", solana_service.DefaultBaseURL)
		c.SolanaService = solana_service.New(rest.New(solana_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: solana_service.Service, Field: "solanaService", EnvVar: "SOLANA_SERVICE_URL", BaseURL: u, Endpoints: 15})
	}
	{
		u := baseURL("THREADS_SERVICE_URL", threads_service.DefaultBaseURL)
		c.ThreadsService = threads_service.New(rest.New(threads_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: threads_service.Service, Field: "threadsService", EnvVar: "THREADS_SERVICE_URL", BaseURL: u, Endpoints: 8})
	}
	{
		u := baseURL("TICKET_SERVICE_URL", ticket_service.DefaultBaseURL)
		c.TicketService = ticket_service.New(rest.New(ticket_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: ticket_service.Service, Field: "ticketService", EnvVar: "TICKET_SERVICE_URL", BaseURL: u, Endpoints: 34})
	}
	{
		u := baseURL("TOGGLE_SERVICE_URL", toggle_service.DefaultBaseURL)
		c.ToggleService = toggle_service.New(rest.New(toggle_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: toggle_service.Service, Field: "toggleService", EnvVar: "TOGGLE_SERVICE_URL", BaseURL: u, Endpoints: 11})
	}
	{
		u := baseURL("TOOL_SERVICE_URL", tool_service.DefaultBaseURL)
		c.ToolService = tool_service.New(rest.New(tool_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: tool_service.Service, Field: "toolService", EnvVar: "TOOL_SERVICE_URL", BaseURL: u, Endpoints: 7})
	}
	{
		u := baseURL("UPLOAD_SERVICE_URL", upload_service.DefaultBaseURL)
		c.UploadService = upload_service.New(rest.New(upload_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: upload_service.Service, Field: "uploadService", EnvVar: "UPLOAD_SERVICE_URL", BaseURL: u, Endpoints: 2})
	}
	{
		u := baseURL("USER_SERVICE_URL", user_service.DefaultBaseURL)
		c.UserService = user_service.New(rest.New(user_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: user_service.Service, Field: "userService", EnvVar: "USER_SERVICE_URL", BaseURL: u, Endpoints: 3})
	}
	{
		u := baseURL("VENDING_MACHINE_SERVICE_URL", vending_machine_service.DefaultBaseURL)
		c.VendingMachineService = vending_machine_service.New(rest.New(vending_machine_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: vending_machine_service.Service, Field: "vendingMachineService", EnvVar: "VENDING_MACHINE_SERVICE_URL", BaseURL: u, Endpoints: 14})
	}
	{
		u := baseURL("VIDEO_SERVICE_URL", video_service.DefaultBaseURL)
		c.VideoService = video_service.New(rest.New(video_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: video_service.Service, Field: "videoService", EnvVar: "VIDEO_SERVICE_URL", BaseURL: u, Endpoints: 3})
	}
	{
		u := baseURL("VOUCHER_SERVICE_URL", voucher_service.DefaultBaseURL)
		c.VoucherService = voucher_service.New(rest.New(voucher_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: voucher_service.Service, Field: "voucherService", EnvVar: "VOUCHER_SERVICE_URL", BaseURL: u, Endpoints: 9})
	}
	{
		u := baseURL("WALLET_SERVICE_URL", wallet_service.DefaultBaseURL)
		c.WalletService = wallet_service.New(rest.New(wallet_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: wallet_service.Service, Field: "walletService", EnvVar: "WALLET_SERVICE_URL", BaseURL: u, Endpoints: 2})
	}
	{
		u := baseURL("WEBRTC_SERVICE_URL", webrtc_service.DefaultBaseURL)
		c.WebrtcService = webrtc_service.New(rest.New(webrtc_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: webrtc_service.Service, Field: "webrtcService", EnvVar: "WEBRTC_SERVICE_URL", BaseURL: u, Endpoints: 2})
	}
	{
		u := baseURL("WISHLISTS_SERVICE_URL", wishlists_service.DefaultBaseURL)
		c.WishlistsService = wishlists_service.New(rest.New(wishlists_service.Service, u, httpClient))
		c.Services = append(c.Services, ServiceInfo{Name: wishlists_service.Service, Field: "wishlistsService", EnvVar: "WISHLISTS_SERVICE_URL", BaseURL: u, Endpoints: 2})
	}
	return c
}
