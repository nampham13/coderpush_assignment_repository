# Product Specification

## Goal

Nhân viên có thể bắt đầu một catalog update, xem tiến độ/thay
đổi/issues, và resume update dở dang sau khi app restart. Dữ liệu demo
hoàn toàn giả lập.

## Scope

**Có:** Go HTTP API, SQLite, JSON fixtures, batches, validation,
create/update/withdraw/reactivate, version ordering, persistent
issues/history, retry/resume, tests, demo. **Không có trong MVP:**
supplier thật, auth, frontend hoàn chỉnh, distributed workers, parallel
batch processing, tự suy luận withdrawal từ product bị thiếu.

## Event schema

``` json
{
  "event_id": "evt-001",
  "product_id": "SKU-100",
  "source_version": 3,
  "action": "UPSERT",
  "name": "Widget",
  "description": "Fictional item",
  "price_minor": 1299,
  "currency": "USD",
  "stock_quantity": 15
}
```

-   `event_id`: định danh ổn định.
-   `product_id`: trim hai đầu, không rỗng, case-sensitive.
-   `source_version`: integer dương, tăng đơn điệu theo product.
-   `action`: `UPSERT`, `WITHDRAW`, `REACTIVATE`.
-   `UPSERT`/`REACTIVATE`: name không rỗng, price_minor \>= 0, currency
    gồm 3 chữ in hoa, stock_quantity \>= 0. Description mặc định chuỗi
    rỗng.
-   `WITHDRAW` chỉ cần event_id, product_id, version, action.
-   Unknown action, missing field, malformed types và giá trị không hợp
    lệ tạo persistent issue; không làm sập cả update.

## Invariants bắt buộc

1.  Mỗi product_id có tối đa một product.
2.  Version cũ không ghi đè version mới.
3.  Event cũ không thể khôi phục sản phẩm đã withdrawn.
4.  Chỉ REACTIVATE hợp lệ với version cao hơn mới kích hoạt lại.
5.  Replay không tạo duplicate mutation, receipt, history hoặc issue.
6.  Product mutation, event outcome và history được commit trong cùng
    transaction.
7.  Update không complete nếu còn batch chưa xử lý hoặc lỗi theo policy
    chưa kết thúc.
8.  Invalid event không làm mất catalog hiện có.
9.  Không có product trong batch không đồng nghĩa với withdrawal.
10. Snapshot input phải bất biến giữa retry/resume.
11. Cùng version nhưng payload/action khác là `VERSION_CONFLICT`; giữ
    state hiện tại.
12. Cùng event_id nhưng payload khác là `EVENT_ID_CONFLICT`; không ghi
    đè event gốc.

## Version policy

-   Product chưa tồn tại + UPSERT: tạo mới.
-   Version lớn hơn hiện tại: áp dụng action hợp lệ.
-   Version nhỏ hơn: `STALE_IGNORED`, không mutate.
-   Cùng version và cùng payload/action: `DUPLICATE_NOOP`.
-   Cùng version nhưng payload/action khác: issue conflict, giữ state.
-   WITHDRAW giữ tombstone và source_version.
-   REACTIVATE khi product chưa từng tồn tại: reject trong MVP.
-   So sánh source_version theo product; không dùng local timestamp để
    quyết định ordering.

## States

Batch: `PENDING`, `RUNNING`, `COMPLETED`, `COMPLETED_WITH_ISSUES`,
`RETRYABLE_FAILED`, `FAILED`. Update: `PENDING`, `RUNNING`, `COMPLETED`,
`COMPLETED_WITH_ISSUES`, `INTERRUPTED`, `FAILED`. `CANCELLED` chỉ thêm
nếu có cancel API. - `COMPLETED`: tất cả batch xong, không có
issue/error. - `COMPLETED_WITH_ISSUES`: mọi batch đã được xử lý đến
trạng thái cuối, nhưng có event rejected/conflict. - `INTERRUPTED`: còn
việc có thể resume. - `FAILED`: lỗi permanent/không thể tiếp tục. - Chỉ
một active update tại một thời điểm; enforce bằng transaction, không chỉ
bằng UI. - Retry lỗi transient; validation/conflict không tự retry vô
hạn. - Batch chỉ complete sau khi mọi event outcome đã commit.

## Progress and issues

Mỗi event có outcome: `CREATED`, `UPDATED`, `WITHDRAWN`, `REACTIVATED`,
`UNCHANGED`, `DUPLICATE_NOOP`, `STALE_IGNORED`, `REJECTED`,
`VERSION_CONFLICT`, `EVENT_ID_CONFLICT`. Lưu issue ID,
update/batch/event/product association, code, severity, message,
OPEN/RESOLVED, timestamps và dedupe key. Progress phải lấy từ persistent
state, không đếm lặp khi retry.

## Acceptance criteria

-   Newer version wins; stale version ignored.
-   Duplicate events do not duplicate products/history.
-   Withdrawal cannot be undone by stale UPSERT.
-   Invalid records create issues while valid records proceed.
-   Restart/resume preserves committed work.
-   Double start cannot create two active updates.
-   Update never claims completion while batches remain.
-   API exposes progress, products, history and issues.
-   README documents assumptions, run/test steps and limitations.
