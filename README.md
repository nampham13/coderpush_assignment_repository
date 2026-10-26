# Supplier Catalog Updates --- Copilot Specification Pack

Bộ spec cho ứng dụng Go nhập catalog từ supplier giả lập. Đây là tài
liệu triển khai, chưa phải source code của ứng dụng.

## Đọc theo thứ tự

1.  `docs/SPEC.md` --- yêu cầu nghiệp vụ và invariants.
2.  `docs/ARCHITECTURE.md` --- cấu trúc module và luồng xử lý.
3.  `docs/DATA_MODEL.md` --- database và transaction.
4.  `docs/API_SPEC.md` --- HTTP API.
5.  `docs/TEST_SPEC.md` --- test matrix và Definition of Done.
6.  `docs/IMPLEMENTATION_PLAN.md` --- task queue theo thứ tự.
7.  `.github/copilot-instructions.md` --- chỉ dẫn repository-wide.
8.  `.github/prompts/implement-task.prompt.md` --- prompt triển khai
    từng task.

## Baseline

-   Go, HTTP JSON API, SQLite, JSON fixtures giả lập supplier.
-   Xử lý tuần tự trong MVP; chỉ một update hoạt động tại một thời điểm.
-   Product ID là identity; version do supplier cung cấp quyết định thứ
    tự.
-   Withdrawal lưu tombstone, không xóa cứng.
-   Input snapshot bất biến; trạng thái/tiến độ lưu bền vững.
-   Product mutation, event outcome và history phải nhất quán
    transactionally.

## Dùng với GitHub Copilot

Mở root repository trong IDE, chọn prompt file trong `.github/prompts/`
nếu client hỗ trợ. Giao một task ID mỗi lần; review diff và chạy test
trước khi sang task kế tiếp. Hỗ trợ custom instructions/prompt files có
thể khác nhau theo IDE và phiên bản.

## Tài liệu GitHub chính thức

-   [Custom
    instructions](https://docs.github.com/en/copilot/concepts/prompting-and-context/custom-instructions)
-   [Repository instructions and prompt
    files](https://docs.github.com/en/copilot/how-tos/copilot-on-github/customize-copilot/add-custom-instructions/add-repository-instructions)
-   [Customization cheat
    sheet](https://docs.github.com/en/copilot/reference/customization-cheat-sheet)
-   [Best practices for coding
    tasks](https://docs.github.com/en/copilot/using-github-copilot/using-copilot-coding-agent-to-work-on-tasks/best-practices-for-using-copilot-to-work-on-tasks)
