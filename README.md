# Codebase Exporter

## Mục đích

Codebase Exporter là công cụ giúp xuất toàn bộ mã nguồn từ một thư mục (codebase) thành một file txt đơn lẻ, được thiết kế đặc biệt để tạo ra prompt hiệu quả khi làm việc với các mô hình ngôn ngữ lớn (LLM) như ChatGPT, Claude, hay Gemini.

Công cụ này giúp bạn:

- Tạo cái nhìn tổng quan về cấu trúc thư mục dự án
- Trích xuất nội dung các file mã nguồn theo định dạng dễ đọc
- Lọc các loại file cần thiết dựa trên phần mở rộng
- Tổng hợp thành một file duy nhất để có thể dễ dàng sao chép và dán vào prompt cho AI

## Tính năng

- Giao diện đồ họa dễ sử dụng
- Chọn thư mục dự án nguồn và thư mục đầu ra
- Lọc file theo phần mở rộng (mặc định: cpp, h, txt)
- Hiển thị cấu trúc thư mục dạng cây
- Tạo file output.txt chứa toàn bộ mã nguồn với định dạng có tổ chức

## Cách sử dụng

1. Khởi động ứng dụng
2. Nhấn nút "Browse..." để chọn thư mục dự án cần xuất
3. Nhấn nút "Browse..." thứ hai để chọn thư mục đầu ra
4. Nhập các phần mở rộng file cần bao gồm, cách nhau bằng dấu phẩy (ví dụ: cpp,h,txt,py)
   - Để trống nếu muốn xuất tất cả các file
5. Nhấn nút "Process Directory" để bắt đầu xử lý
6. Chờ đến khi thanh trạng thái hiển thị thông báo hoàn thành
7. File output.txt sẽ được tạo trong thư mục đầu ra đã chọn

## Gợi ý sử dụng với AI

1. Chạy Codebase Exporter và tạo file src.txt
2. Mở file output.txt và chỉnh sửa phần "project overview" hoặc thêm context theo ý bạn, sau đó sao chép file.
3. Dán nội dung vào prompt cho AI với yêu cầu phân tích cụ thể
4. Do file bao gồm cấu trúc thư mục và mã nguồn, AI có thể hiểu được bối cảnh và mối quan hệ giữa các file

## Lưu ý

- Hạn chế sử dụng với các dự án quá lớn vì giới hạn token của các AI
- Đảm bảo bạn có quyền truy cập vào tất cả các thư mục trong dự án

## Hướng dẫn build từ mã nguồn

```bash
go build -ldflags="-H windowsgui"
go build -o codebase-exporter .
```
