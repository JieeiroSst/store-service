from __future__ import annotations

from typing import Dict, List

from app.domain.conversation import DEFAULT_SCENARIO
from app.domain.models import KnowledgeEntry, Scenario

DEFAULT_SCENARIOS: Dict[str, Dict[str, str]] = {
    "customer_support": {
        "name": "Hỗ trợ khách hàng",
        "greeting": "Xin chào! Tôi là trợ lý ảo của tổng đài chăm sóc khách hàng. Tôi có thể giúp gì cho bạn?",
        "prompt": """Bạn là nhân viên tổng đài chăm sóc khách hàng chuyên nghiệp, thân thiện và kiên nhẫn.
Nhiệm vụ:
- Xác định nhanh nhu cầu: hỏi về đơn hàng, giao hàng, đổi trả, thanh toán, bảo hành, khuyến mãi, tài khoản hay khiếu nại.
- Trả lời đúng trọng tâm bằng kiến thức tham khảo, rồi hỏi khách còn cần hỗ trợ gì thêm không.
- Với câu hỏi về một đơn hàng cụ thể, hướng dẫn khách cách tự tra cứu; nếu khách vẫn cần kiểm tra chi tiết thì chuyển nhân viên.
- Luôn giữ giọng điệu lịch sự, xưng "tôi" và gọi khách là "bạn" hoặc theo tên nếu biết.""",
    },
    "order_support": {
        "name": "Hỗ trợ đơn hàng và đổi trả",
        "greeting": "Xin chào! Đây là tổng đài hỗ trợ đơn hàng. Bạn cần kiểm tra đơn, đổi trả hay hoàn tiền ạ?",
        "prompt": """Bạn là nhân viên hỗ trợ đơn hàng, giao nhận, đổi trả và hoàn tiền.
Nhiệm vụ:
- Hỏi khách đang cần: kiểm tra trạng thái đơn, thay đổi hoặc hủy đơn, đổi trả, hay hoàn tiền.
- Giải thích điều kiện, thời hạn và các bước theo đúng chính sách trong kiến thức tham khảo.
- Khi khách cần đổi trả, nhắc khách chuẩn bị mã đơn hàng, giữ nguyên tem nhãn và hóa đơn.
- Bạn không tra cứu được hệ thống đơn hàng; nếu khách cần xác nhận tình trạng cụ thể của một đơn thì xin mã đơn hàng để nhân viên kiểm tra và chuyển nhân viên.""",
    },
    "sales": {
        "name": "Tư vấn bán hàng",
        "greeting": "Xin chào! Tôi là trợ lý tư vấn mua sắm. Bạn đang quan tâm sản phẩm hay chương trình khuyến mãi nào ạ?",
        "prompt": """Bạn là nhân viên tư vấn bán hàng chuyên nghiệp, nhiệt tình nhưng không ép mua.
Nhiệm vụ:
- Tìm hiểu nhu cầu, ngân sách và mục đích sử dụng của khách bằng từng câu hỏi ngắn.
- Giới thiệu lợi ích, chương trình khuyến mãi, phương thức thanh toán và giao hàng dựa trên kiến thức tham khảo.
- Xử lý băn khoăn về giá hoặc chất lượng một cách khéo léo, trung thực; không hứa hẹn điều không có trong kiến thức tham khảo.
- Khi khách muốn đặt hàng hoặc cần báo giá chi tiết, hướng dẫn kênh đặt hàng hoặc chuyển nhân viên kinh doanh.""",
    },
    "complaint": {
        "name": "Tiếp nhận khiếu nại",
        "greeting": "Xin chào! Đây là bộ phận tiếp nhận phản ánh. Bạn vui lòng cho biết vấn đề bạn gặp phải ạ.",
        "prompt": """Bạn là nhân viên tiếp nhận khiếu nại, bình tĩnh và đồng cảm.
Nhiệm vụ:
- Lắng nghe, xin lỗi về trải nghiệm chưa tốt và tóm tắt lại vấn đề để xác nhận đã hiểu đúng.
- Thu thập lần lượt: mã đơn hàng hoặc thời điểm phát sinh, mô tả vấn đề, mong muốn của khách.
- Nêu quy trình và thời hạn xử lý khiếu nại theo kiến thức tham khảo.
- Không tranh cãi, không đổ lỗi. Nếu khách rất bức xúc hoặc vấn đề nghiêm trọng thì chuyển nhân viên ngay.""",
    },
    "appointment": {
        "name": "Đặt lịch hẹn",
        "greeting": "Xin chào! Tôi là trợ lý đặt lịch hẹn. Bạn muốn đặt lịch vào ngày nào ạ?",
        "prompt": """Bạn là trợ lý đặt lịch hẹn tự động.
Nhiệm vụ:
- Thu thập lần lượt: họ tên, số điện thoại liên hệ, dịch vụ cần hẹn, ngày giờ mong muốn.
- Chỉ nhận lịch trong giờ làm việc ghi ở thông tin công ty.
- Đọc lại toàn bộ thông tin để khách xác nhận, sau đó thông báo lịch hẹn đã được ghi nhận và nhân viên sẽ gọi xác nhận lại.""",
    },
    "survey": {
        "name": "Khảo sát ý kiến",
        "greeting": "Xin chào! Chúng tôi xin phép dành một phút để hỏi ý kiến của bạn về dịch vụ. Bạn có đồng ý tham gia không ạ?",
        "prompt": """Bạn là nhân viên khảo sát mức độ hài lòng của khách hàng.
Nhiệm vụ:
- Hỏi lần lượt, mỗi lượt một câu: mức độ hài lòng từ 1 đến 5, điều khách thích nhất, điều cần cải thiện, khả năng giới thiệu cho người quen.
- Ghi nhận câu trả lời, không tranh luận. Nếu khách từ chối thì cảm ơn và kết thúc.
- Cảm ơn khách vì đã dành thời gian khi hoàn tất.""",
    },
}

DEFAULT_KNOWLEDGE: List[Dict[str, str]] = [
    {
        "category": "chung",
        "question": "Giờ làm việc của tổng đài và cửa hàng là khi nào?",
        "answer": "Tổng đài hỗ trợ từ 8 giờ sáng đến 21 giờ tối tất cả các ngày trong tuần, kể cả cuối tuần. Ngày lễ Tết có thể thay đổi và sẽ được thông báo trên website.",
        "keywords": "giờ làm việc, mấy giờ, mở cửa, đóng cửa, thời gian hỗ trợ, cuối tuần, ngày lễ",
    },
    {
        "category": "giao hàng",
        "question": "Phí giao hàng là bao nhiêu?",
        "answer": "Đơn hàng từ 300 nghìn đồng được miễn phí giao hàng tiêu chuẩn. Đơn dưới 300 nghìn tính phí từ 20 đến 35 nghìn đồng tùy khu vực, phí chính xác hiển thị ở bước thanh toán.",
        "keywords": "phí ship, phí vận chuyển, miễn phí giao hàng, freeship, tiền ship",
    },
    {
        "category": "giao hàng",
        "question": "Bao lâu thì nhận được hàng?",
        "answer": "Nội thành Hà Nội và Thành phố Hồ Chí Minh giao trong 1 đến 2 ngày làm việc. Các tỉnh thành khác từ 3 đến 5 ngày làm việc. Có giao hỏa tốc trong 4 giờ ở nội thành với phụ phí.",
        "keywords": "thời gian giao hàng, bao lâu nhận hàng, mấy ngày, giao nhanh, hỏa tốc, khi nào nhận",
    },
    {
        "category": "đơn hàng",
        "question": "Làm sao để kiểm tra tình trạng đơn hàng?",
        "answer": "Bạn có thể kiểm tra trong mục Đơn hàng của tôi trên ứng dụng hoặc website, hoặc theo đường link theo dõi gửi qua tin nhắn khi đơn được giao cho đơn vị vận chuyển. Nếu cần nhân viên kiểm tra, bạn vui lòng cung cấp mã đơn hàng.",
        "keywords": "kiểm tra đơn, tra cứu đơn, đơn hàng đâu, tình trạng đơn, theo dõi đơn, mã vận đơn, chưa nhận được hàng",
    },
    {
        "category": "đơn hàng",
        "question": "Tôi muốn hủy đơn hoặc thay đổi địa chỉ giao hàng thì làm thế nào?",
        "answer": "Bạn có thể tự hủy hoặc sửa địa chỉ trong mục Đơn hàng của tôi khi đơn còn ở trạng thái chờ xác nhận. Nếu đơn đã được bàn giao cho đơn vị vận chuyển, nhân viên sẽ hỗ trợ liên hệ bên giao hàng.",
        "keywords": "hủy đơn, huỷ đơn, đổi địa chỉ, sửa địa chỉ, thay đổi đơn hàng, đổi số điện thoại nhận hàng",
    },
    {
        "category": "đổi trả",
        "question": "Chính sách đổi trả hàng như thế nào?",
        "answer": "Sản phẩm được đổi trả trong vòng 7 ngày kể từ khi nhận hàng nếu còn nguyên tem nhãn, chưa qua sử dụng và có hóa đơn. Hàng lỗi do nhà sản xuất được đổi miễn phí trong 30 ngày. Bạn tạo yêu cầu đổi trả trong mục Đơn hàng của tôi.",
        "keywords": "đổi trả, trả hàng, đổi hàng, hàng lỗi, đổi size, không vừa, hoàn hàng",
    },
    {
        "category": "đổi trả",
        "question": "Bao lâu thì được hoàn tiền?",
        "answer": "Sau khi kho nhận và kiểm tra hàng trả, tiền được hoàn trong 3 đến 7 ngày làm việc về đúng phương thức thanh toán ban đầu. Với thẻ tín dụng có thể mất đến 14 ngày tùy ngân hàng.",
        "keywords": "hoàn tiền, trả lại tiền, refund, khi nào nhận lại tiền, chưa nhận được tiền hoàn",
    },
    {
        "category": "thanh toán",
        "question": "Có những phương thức thanh toán nào?",
        "answer": "Bạn có thể thanh toán khi nhận hàng, chuyển khoản ngân hàng, thẻ ATM nội địa, thẻ Visa hoặc Mastercard và ví điện tử. Đơn trên 10 triệu đồng có hỗ trợ trả góp qua thẻ tín dụng.",
        "keywords": "thanh toán, trả tiền, cod, chuyển khoản, thẻ, visa, ví điện tử, momo, trả góp",
    },
    {
        "category": "thanh toán",
        "question": "Tôi đã bị trừ tiền nhưng đơn hàng báo chưa thanh toán thì sao?",
        "answer": "Trường hợp này thường được hệ thống tự đối soát trong vòng 24 giờ. Nếu sau 24 giờ đơn vẫn chưa cập nhật, nhân viên sẽ kiểm tra giao dịch cho bạn, bạn vui lòng chuẩn bị mã đơn hàng và thời gian giao dịch.",
        "keywords": "bị trừ tiền, trừ tiền rồi, chưa thanh toán, lỗi thanh toán, giao dịch thất bại, mất tiền",
    },
    {
        "category": "bảo hành",
        "question": "Sản phẩm được bảo hành như thế nào?",
        "answer": "Thời hạn bảo hành ghi trên phiếu bảo hành hoặc trang sản phẩm, thường từ 6 đến 24 tháng. Bạn mang sản phẩm kèm hóa đơn đến trung tâm bảo hành hoặc tạo yêu cầu bảo hành trên website để được hướng dẫn gửi hàng.",
        "keywords": "bảo hành, sửa chữa, hỏng, phiếu bảo hành, trung tâm bảo hành",
    },
    {
        "category": "khuyến mãi",
        "question": "Hiện có chương trình khuyến mãi hay mã giảm giá nào không?",
        "answer": "Các chương trình khuyến mãi và mã giảm giá đang áp dụng được cập nhật trong mục Khuyến mãi trên ứng dụng. Mỗi đơn hàng áp dụng một mã giảm giá và một mã miễn phí vận chuyển.",
        "keywords": "khuyến mãi, giảm giá, voucher, mã giảm giá, coupon, ưu đãi, sale, quà tặng",
    },
    {
        "category": "khuyến mãi",
        "question": "Mã giảm giá của tôi không dùng được?",
        "answer": "Bạn vui lòng kiểm tra hạn sử dụng, giá trị đơn tối thiểu và danh mục sản phẩm áp dụng của mã. Mỗi mã chỉ dùng một lần cho mỗi tài khoản. Nếu vẫn không được, nhân viên sẽ kiểm tra giúp bạn.",
        "keywords": "mã không dùng được, voucher lỗi, mã hết hạn, không áp dụng được mã",
    },
    {
        "category": "tài khoản",
        "question": "Tôi quên mật khẩu tài khoản thì làm sao?",
        "answer": "Bạn chọn Quên mật khẩu ở màn hình đăng nhập và làm theo hướng dẫn gửi qua tin nhắn hoặc email. Vì lý do bảo mật, tổng đài không hỏi và không cấp lại mật khẩu hay mã OTP qua điện thoại.",
        "keywords": "quên mật khẩu, đăng nhập, không đăng nhập được, mật khẩu, otp, tài khoản bị khóa",
    },
    {
        "category": "hóa đơn",
        "question": "Tôi muốn xuất hóa đơn VAT?",
        "answer": "Bạn nhập thông tin xuất hóa đơn ở bước thanh toán hoặc gửi yêu cầu trong vòng 7 ngày kể từ khi nhận hàng. Hóa đơn điện tử sẽ được gửi qua email trong 3 ngày làm việc.",
        "keywords": "hóa đơn, hoá đơn, vat, xuất hóa đơn, hóa đơn đỏ, hóa đơn điện tử",
    },
    {
        "category": "khiếu nại",
        "question": "Quy trình xử lý khiếu nại như thế nào?",
        "answer": "Khiếu nại được ghi nhận ngay và phản hồi trong vòng 24 giờ làm việc. Trường hợp cần xác minh với đơn vị vận chuyển hoặc nhà cung cấp, thời gian xử lý tối đa 5 ngày làm việc và bạn sẽ được cập nhật qua tin nhắn.",
        "keywords": "khiếu nại, phản ánh, không hài lòng, thái độ, giao sai, thiếu hàng, hàng giả",
    },
    {
        "category": "chung",
        "question": "Có thể liên hệ công ty qua những kênh nào?",
        "answer": "Bạn có thể liên hệ qua tổng đài này, khung chat trên website và ứng dụng, hoặc email chăm sóc khách hàng ghi trên website. Khung chat hoạt động cùng giờ với tổng đài.",
        "keywords": "liên hệ, email, chat, fanpage, kênh hỗ trợ, số điện thoại",
    },
    {
        "category": "đặt lịch",
        "question": "Có thể đặt lịch hẹn vào những khung giờ nào?",
        "answer": "Lịch hẹn nhận từ 9 giờ sáng đến 18 giờ chiều, thứ Hai đến thứ Bảy. Mỗi lịch hẹn kéo dài khoảng 30 phút và nhân viên sẽ gọi xác nhận trước một ngày.",
        "keywords": "đặt lịch, lịch hẹn, hẹn gặp, khung giờ, đổi lịch, hủy lịch",
        "scenario": "appointment",
    },
]


def default_scenarios() -> List[Scenario]:
    return [
        Scenario(
            name=key,
            description=data["name"],
            greeting=data["greeting"],
            prompt=data["prompt"],
            is_default=1 if key == DEFAULT_SCENARIO else 0,
        )
        for key, data in DEFAULT_SCENARIOS.items()
    ]


def builtin_scenario(name: str) -> Scenario:
    key = name if name in DEFAULT_SCENARIOS else DEFAULT_SCENARIO
    data = DEFAULT_SCENARIOS[key]
    return Scenario(name=key, description=data["name"], greeting=data["greeting"], prompt=data["prompt"])


def default_knowledge() -> List[KnowledgeEntry]:
    return [
        KnowledgeEntry(
            question=item["question"],
            answer=item["answer"],
            keywords=item.get("keywords"),
            category=item.get("category"),
            scenario=item.get("scenario"),
        )
        for item in DEFAULT_KNOWLEDGE
    ]
