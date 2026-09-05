# BeeBox Auth — Một bản Clerk "may đo", kể từ trên xuống dưới

## Vì sao làm cái này

Clerk, Auth0, Supabase Auth — mấy ông lớn này đều bán cho khách một bộ khung cố định: bạn là "user" thì có email, password, tên, và vài field mở rộng nếu trả thêm tiền. Nhưng thực tế thì mỗi sản phẩm cần một cái user khác nhau. Một app giao hàng cần `phoneNumber` là bắt buộc, `fullName` là đủ. Một app doanh nghiệp cần `firstName`/`lastName` tách riêng vì email hệ thống của họ ghép tên theo kiểu đó, cộng thêm `isVerify` để biết ai đã được HR duyệt. Ép tất cả vào một cái schema chung là ép khách phải "lách luật" bằng cách nhét mọi thứ vào field `metadata` chung chung, rồi tự parse lại — vừa xấu vừa dễ lỗi.

Ý tưởng ở đây là: mỗi project được quyền định nghĩa "user của tôi trông như thế nào", còn hệ thống lo phần còn lại — lưu trữ, xác thực, bảo mật, vận hành. Và thay vì bắt khách cài SDK, tự quản lý config, họ được đưa thẳng một backend chạy sẵn, có link để bấm vào test ngay, cùng một cặp khóa (public key, secret key) để gắn vào ứng dụng thật khi đã ưng ý.

## Bức tranh toàn hệ thống

Nhìn từ trên cao, hệ thống không phải một khối to mà là mấy dịch vụ nhỏ, mỗi cái làm đúng một việc, nói chuyện với nhau qua mạng chứ không đọc chung database của nhau. Đây là lựa chọn có chủ đích để sau này scale — dịch vụ nào bị nghẽn thì nhân bản riêng dịch vụ đó, không phải nhân bản cả hệ thống.

Có bốn mảnh chính:

**Gateway** đứng ngoài cùng, là cửa duy nhất mà internet nhìn thấy. Mọi request — dù là từ app của khách hàng B2B hay từ trang test trên site của mình — đều đi qua đây trước. Gateway không biết gì về nghiệp vụ đăng nhập hay field nào của project nào; việc của nó chỉ là: nhìn request, xác định đây là project nào (qua public key hoặc secret key), rồi chuyển tiếp đến đúng dịch vụ phía sau kèm theo một "ngữ cảnh đáng tin" (project nào, environment nào, ai đang gọi). Nó giống bảo vệ ở cổng — kiểm tra thẻ, biết bạn được vào khu nào, nhưng không tham gia vào việc bạn làm gì bên trong.

**Project Service** là nơi giữ "trí nhớ cấu hình" của toàn hệ thống. Nó biết một project tên gì, thuộc về ai, đang ở tier nào (free hay enterprise), và quan trọng nhất — nó giữ định nghĩa field mà project đó chọn cho user của mình, cùng với toàn bộ lịch sử thay đổi của những định nghĩa đó. Nó cũng là nơi phát hành và thu hồi public key/secret key. Đây là bộ não hành chính, không phải nơi diễn ra chuyện đăng nhập thật.

**Identity Service** là nơi thật sự lưu user cuối và xử lý đăng ký/đăng nhập. Nó không tự quyết định user cần field gì — nó hỏi Project Service "project này cấu hình ra sao", rồi dựa vào đó để nhận, kiểm tra và lưu dữ liệu người dùng gửi lên. Về sau, khi có nhu cầu, nó cũng là nơi phát hành JWT cho phiên đăng nhập.

**Hàng đợi (RabbitMQ)** không phải một service, mà là hệ thần kinh chạy ngầm nối các dịch vụ lại mà không cần chúng gọi trực tiếp nhau. Khi Identity Service tạo xong một user mới, nó không cần tự đi gửi email chào mừng hay tự ghi log — nó chỉ thả một sự kiện "user vừa được tạo" lên hàng đợi, rồi ai cần thì tự lắng nghe. Sau này khi có thêm dịch vụ gửi email, dịch vụ ghi audit log, hay webhook báo cho khách hàng B2B biết "vừa có người đăng ký", tất cả chỉ cần đăng ký lắng nghe cùng một sự kiện, không phải sửa Identity Service.

Ba dịch vụ này đứng riêng, mỗi cái có database, có vòng đời deploy, có khả năng chịu tải riêng — dịch vụ Identity chắc chắn sẽ nhận nhiều traffic hơn Project Service rất nhiều lần (vì user cuối đăng nhập liên tục, còn config project thì thi thoảng mới đổi), nên tách ra để sau này scale đúng chỗ cần scale, không phải kéo theo cả cụm.

## Vì sao dữ liệu user lại trông như vậy

Bên trong Identity Service, một user không được lưu theo kiểu "mỗi project một bảng riêng". Lý do đơn giản: có bao nhiêu project là có bấy nhiêu bảng, và mỗi lần cần sửa một chút ở tầng hạ tầng (thêm index, đổi kiểu dữ liệu) là phải chạy lại trên toàn bộ hàng nghìn bảng đó — vận hành sẽ thành ác mộng rất nhanh.

Thay vào đó, tất cả user cuối nằm chung một bảng. Những gì bắt buộc phải có ở mọi user — email để đăng nhập, mật khẩu đã băm, project nào sở hữu — nằm ở cột riêng, vì đây là những thứ cần tra cứu nhanh và cần ràng buộc chặt (không được trùng email trong cùng một project, chẳng hạn). Còn tất cả những field mà từng project tự nghĩ ra — `fullName`, `firstName`, `isVerify`, hay bất cứ thứ gì họ muốn — được gom vào một cột duy nhất chứa dữ liệu dạng JSON. Cột này linh hoạt, không đòi hỏi phải sửa cấu trúc bảng mỗi khi có project mới với nhu cầu mới.

Nhưng linh hoạt không có nghĩa là buông lỏng. Mỗi project có một danh sách mô tả rõ ràng: field này tên gì, kiểu dữ liệu gì, có bắt buộc không, có phải thông tin nhạy cảm không. Danh sách này sống ở Project Service, và mọi dữ liệu gửi lên trước khi được ghi vào cột JSON kia đều phải đi qua bước kiểm tra dựa trên chính danh sách đó. Nói cách khác: sự tự do của khách hàng B2B trong việc "muốn field gì cũng được" không đồng nghĩa với việc hệ thống chấp nhận rác — nó chỉ có nghĩa là luật được định nghĩa động, thay vì đóng cứng trong code.

Có một chi tiết dễ bị bỏ qua nhưng lại quan trọng: khách hàng B2B sẽ đổi ý. Hôm nay họ dùng `fullName`, sáu tháng sau họ muốn tách thành `firstName`/`lastName`. Nếu sửa thẳng vào danh sách field hiện tại, toàn bộ user cũ đã có `fullName` trong dữ liệu JSON của họ bỗng dưng không khớp với "luật mới" nữa — dễ gây lỗi âm thầm rất khó phát hiện. Vì vậy mỗi lần danh sách field của một project thay đổi về cấu trúc, hệ thống lưu lại một phiên bản mới, và mỗi user được gắn với đúng phiên bản đang có hiệu lực tại thời điểm họ được tạo. Đọc dữ liệu của một user luôn dựa vào phiên bản gắn trên chính user đó, không phải phiên bản mới nhất — user cũ không bao giờ bị "diễn giải sai" chỉ vì khách hàng B2B đổi cấu hình sau này.

## Chuyện co giãn theo từng khách hàng

Không phải khách hàng nào cũng giống nhau. Một startup nhỏ dùng gói miễn phí thì nằm chung database với hàng trăm project khác cũng chẳng sao — chi phí thấp, vận hành đơn giản, đủ dùng. Nhưng một khách hàng lớn, làm trong ngành tài chính hay y tế, có thể yêu cầu dữ liệu của họ phải nằm tách biệt hoàn toàn vì lý do tuân thủ, không được phép chung ổ đĩa với ai khác.

Thay vì coi đây là hai hệ thống khác nhau, Identity Service được xây sao cho "nơi lưu dữ liệu" là một chi tiết có thể thay thế, không phải thứ khắc chết vào logic nghiệp vụ. Ở tầng xử lý, sign-up hay sign-in không cần biết dữ liệu đang nằm ở đâu — nó chỉ gọi "lưu user này giúp tôi" thông qua một lớp trung gian, và lớp đó mới là nơi quyết định: project này thuộc tier thường thì ghi vào database dùng chung, project kia thuộc tier cao thì ghi vào một database riêng đã được cấp cho họ. Ngày đầu tiên, có lẽ chỉ cần một cách lưu duy nhất — dùng chung. Nhưng vì lớp trung gian này đã tồn tại từ đầu như một ranh giới rõ ràng, ngày nào đó cần thêm cách lưu riêng cho khách VIP, đó là chuyện thêm một mảnh mới, không phải mổ xẻ lại toàn bộ logic đăng ký/đăng nhập đã chạy ổn định.

## Một khách hàng B2B trải nghiệm hệ thống này như thế nào

Hãy tưởng tượng một người tên là Minh, đang làm một app giao đồ ăn nhỏ, cần chức năng đăng nhập nhưng không muốn tự viết. Minh vào trang của mình, tạo một project mới, đặt tên "GiaoDoNhanh". Ngay sau đó, Minh được hỏi: user của bạn cần những thông tin gì? Minh chọn `fullName` (bắt buộc), `phoneNumber` (bắt buộc), `email` (luôn có sẵn, không cần chọn). Xong xuôi, hệ thống đưa cho Minh một đường link test riêng cho project này, cùng một cặp khóa — một khóa để dán vào phần frontend của app (an toàn khi lộ ra), một khóa để dùng ở phần server của Minh (tuyệt đối giữ kín).

Minh có thể bấm ngay vào link đó, thử đăng ký một tài khoản giả, xem form hiện ra đúng ba field mình vừa chọn hay không, thử nhập sai để xem lỗi trả về có rõ ràng không — tất cả mà chưa cần viết một dòng code nào. Khi ưng ý, Minh gắn hai khóa đó vào app thật của mình, và từ giờ mọi lần người dùng của Minh bấm "Đăng ký" trên app "GiaoDoNhanh", request đó sẽ chạy qua đúng cấu hình mà Minh đã chọn.

Sáu tháng sau, Minh quyết định muốn thêm field `referralCode` để làm chương trình giới thiệu bạn bè. Minh vào lại cấu hình project, thêm field mới. Những người dùng cũ không có `referralCode` không hề bị ảnh hưởng gì — dữ liệu của họ vẫn đọc đúng như lúc họ đăng ký, còn người dùng mới từ giờ sẽ được hỏi thêm field đó.

## Còn người dùng cuối — người thật sự bấm nút "Đăng ký" — thì sao

Người dùng cuối của Minh không biết và không cần biết có BeeBox Auth đứng phía sau. Với họ, mọi thứ vẫn là màn hình đăng ký của app "GiaoDoNhanh" — nhập tên, số điện thoại, email, đặt mật khẩu, bấm nút. Cái form đó có thể do chính Minh tự code (gọi thẳng vào API), hoặc — nếu về sau hệ thống có thêm một trang đăng nhập dựng sẵn — Minh chỉ cần trỏ app của mình tới một đường link, và trang đó tự động hiện đúng ba (hay bốn, sau khi có thêm referral code) field mà Minh đã cấu hình, không cần Minh tự vẽ giao diện.

Ở phía sau cánh gà, khi người dùng đó bấm "Đăng ký", request đi qua Gateway trước — Gateway nhìn vào khóa được gửi kèm, biết ngay đây là request của project "GiaoDoNhanh", không phải của ai khác, rồi chuyển tới Identity Service. Identity Service hỏi nhanh Project Service "field hiện tại của project này gồm những gì, cái nào bắt buộc", kiểm tra dữ liệu người dùng vừa gửi có khớp không, rồi mới lưu xuống. Toàn bộ chuyện này diễn ra trong chớp mắt, và người dùng cuối chỉ thấy một dòng chữ quen thuộc: "Đăng ký thành công".

## Thứ tự nên đi

Không cần dựng đủ bốn dịch vụ cùng lúc. Nên bắt đầu với một lõi tối thiểu chạy được hết một luồng thật: Project Service quản lý project và field định nghĩa (có tính phiên bản ngay từ đầu, vì thêm sau sẽ đắt hơn nhiều), Identity Service xử lý đăng ký/đăng nhập dựa trên field đó, Gateway đứng trước hai cái này để định tuyến và xác thực khóa. RabbitMQ có thể chưa cần bật ngay ngày đầu — chờ đến khi có việc thật sự cần xử lý bất đồng bộ (gửi email, ghi log sự kiện) thì thêm vào, tránh dựng hạ tầng cho một nhu cầu chưa tồn tại. Trang đăng nhập dựng sẵn và khả năng tách database riêng cho khách VIP là những thứ nên để dành thiết kế sẵn chỗ đứng (một interface, một cờ cấu hình) nhưng chưa cần xây thật — xây khi có khách hàng thật sự cần đến.