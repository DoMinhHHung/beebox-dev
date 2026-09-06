# Testing matrix — beebox-project

Cập nhật lần cuối: sau Step 10.

## Đã có bằng chứng (test pass)

| Loại | Package | Ghi chú |
|---|---|---|
| Unit — domain validation | `domain/project`, `domain/credential`, `domain/fielddefinition`, `domain/owner`, `domain/ownersession` | Toàn bộ rule validate, hash, versioning |
| Unit — application logic | `application/*` | Constructor injection, dùng memory repo |
| Integration — Postgres CRUD | `infrastructure/*/postgres` | Cần `DATABASE_URL`, tự skip nếu thiếu |
| Security — secret không lộ | `transport/http` (`TestGetCredential_DoesNotLeakSecret`) | Response GET/Revoke tách struct riêng, không có field secret |
| Security — chống email enumeration | `application/auth` (`TestSignIn_UnknownEmail`) | Trả cùng lỗi cho email sai/không tồn tại |
| Security — cross-tenant (đơn lẻ) | `application/credential`, `application/fielddefinition` (`*_WrongOwnerDenied`) | Actor chưa từng sở hữu project nào |
| Security — cross-tenant (owner có project riêng) | `transport/http` (`TestCrossTenant_*`) | Actor có project thật, thử chéo sang project khác — kịch bản tấn công thực tế hơn |
| Concurrency — data race | Toàn bộ (`go test -race`) | Bắt race điều kiện bộ nhớ, không bắt lost-update logic |
| Concurrency — lost update thật | `application/credential` (`TestService_Rotate_ConcurrentCalls_*`) | Xác nhận hành vi hiện tại (xem "Rủi ro đã biết" bên dưới) |
| Hạ tầng — host không tới được | `infrastructure/postgres` (`TestNewPool_UnreachableHost_*`) | Xác nhận có timeout, không treo |
| Hạ tầng — pool đã đóng | `infrastructure/postgres` (`TestPool_AfterClose_*`) | Xác nhận trả lỗi, không panic |

## Rủi ro đã biết, cố tình CHƯA sửa trong Step 10

- **`credential.Service.Rotate` không có optimistic locking.** 2 request `Rotate` đồng thời trên cùng 1 credential ID: cả 2 đều thành công ở tầng Go, nhưng chỉ dòng ghi cuối cùng vào Postgres còn hiệu lực — secret trả về cho lời gọi "thua" sẽ không dùng được. Không phải lỗ hổng cross-tenant (không ai truy cập được dữ liệu người khác), chỉ là vấn đề correctness khi có double-click/retry đồng thời trên cùng 1 credential. Cách sửa khi cần: thêm cột `version`/dùng `updated_at` làm điều kiện `WHERE` khi `Save`, coi 0 row affected là conflict — đây là thay đổi hành vi, cần làm ở 1 step riêng, không lẫn vào Step 10 (Testing).

## Chưa làm (ngoài phạm vi Step 10)

- Test tải (load test) thật sự — không nằm trong "testing matrix" ban đầu, chỉ cần khi chuẩn bị production.
- `testcontainers-go` để chạy Postgres test không phụ thuộc Supabase thật — dự kiến làm ở Step 11 (CI) nếu cần chạy test postgres trong CI mà không có secret `DATABASE_URL` của Supabase.