Bạn là một senior Go engineer. Hãy tạo cho tôi một **project Golang hoàn chỉnh** (CLI tool) với yêu cầu chi tiết như sau.

## Mục tiêu tổng quan

Viết một chương trình Go dạng CLI, dùng để:

- Đọc danh sách các Go module từ một file đầu vào (ví dụ `modules.txt`), mỗi dòng có dạng:
  - `github.com/gin-gonic/gin@v1.9.1` hoặc
  - `github.com/golang/protobuf` (không có @version thì hiểu là lấy version mới nhất mà `go` resolve được)
- Tự động tải đầy đủ:
  - Các module trong file
  - Toàn bộ module phụ thuộc (dependencies) của chúng (đệ quy)
- Đóng gói toàn bộ các module này theo **định dạng Athens disk storage**, để có thể copy thư mục này sang một máy offline và chạy Athens ở chế độ `StorageType = "disk"`.

## Yêu cầu về Athens disk storage

Chương trình phải sinh ra output đúng cấu trúc mà Athens disk storage cần. Với mỗi module `M` và version `V`, cần có thư mục:

`<storage_root>/<module_path>/<version>/`

<storage_root> mặc định là thư mục repos trong thư mục hiện tại.

Trong thư mục đó phải có tối thiểu 3 file:

- `go.mod`
- `V.info`  (ví dụ: `v1.9.1.info`)
- `source.zip` (zip chứa code của module ở version đó)

Trong đó:

- `<storage_root>` là một thư mục gốc dùng để cấu hình Athens (ví dụ: `/data/athens-storage`).
- `module_path` là path đầy đủ của Go module (ví dụ: `github.com/gin-gonic/gin`).
- `version` là semver của module (bắt đầu bằng `v`, ví dụ `v1.9.1`).

Project **không bắt buộc** phải dùng `pacmod`, nhưng nếu dùng thì phải wrap nó chuẩn để không phụ thuộc quá nhiều vào shell script. Nếu bạn thấy hợp lý, có thể implement logic pack `.info`, `.zip`, `go.mod` bằng Go thuần.

## Yêu cầu chức năng chi tiết

1. **Input modules list**
   - Chương trình nhận tham số `--modules` (hoặc `-m`) để chỉ ra file chứa danh sách module.
   - Mỗi dòng trong file:
     - Bỏ qua dòng trống.
     - Bỏ qua dòng bắt đầu bằng `#` (comment).
     - Cho phép dạng có/không có `@version`.
   - Ví dụ `modules.txt`:
     ```text
     github.com/gin-gonic/gin@v1.9.1
     github.com/sirupsen/logrus@v1.9.3
     # Không chỉ version -> lấy latest
     github.com/golang/protobuf
     ```

2. **Download modules + dependencies (đệ quy)**
   - Chương trình sử dụng cơ chế Go Modules để:
     - Resolve full dependency graph gồm toàn bộ module + version thực tế (`go list -m all` hoặc tương đương bằng Go code).
   - Mục tiêu:
     - Từ danh sách ban đầu, sau khi resolve sẽ có được:
       - `Path` (module path)
       - `Version`
       - `Dir` (thư mục local chứa mã nguồn module trong module cache)
   - Có thể chấp nhận gọi trực tiếp các lệnh `go` (như `go mod init`, `go get`, `go list -m -json all`) thông qua `os/exec`, nhưng cần được wrap cẩn thận, logging rõ ràng.

3. **Sinh output theo format Athens**
   - Với mỗi module `Path` và `Version` (bỏ qua module main, chỉ xử lý module phụ thuộc và module list input):
     - Bỏ qua:
       - Module không có `Version`.
       - Version không phải dạng semver bắt đầu bằng `v` (ví dụ: `devel` hoặc rỗng).
     - Lấy thư mục mã nguồn `Dir`.
   - Từ `Dir`, chương trình phải tạo ra 3 file:
     - `go.mod`
     - `<Version>.info`
     - `source.zip`
   - Sau đó di chuyển/ghi chúng vào thư mục đích:
     - `${storage_root}/${Path}/${Version}/go.mod`
     - `${storage_root}/${Path}/${Version}/${Version}.info`
     - `${storage_root}/${Path}/${Version}/source.zip`
   - Nếu module đã tồn tại đầy đủ trong Athens storage (ít nhất `source.zip` là đủ để coi là done), thì **bỏ qua, không build lại** (idempotent).

   ### Yêu cầu nội dung file `.info`
   - File `<Version>.info` nên tuân theo format JSON cơ bản mà Go proxy dùng, ví dụ (thu gọn):
     ```json
     {
       "Version": "v1.9.1",
       "Time": "2024-01-01T00:00:00Z"
     }
     ```
   - Nếu không dễ lấy chính xác thời gian thực tế của version, có thể:
     - Lấy từ metadata của `go list -m -json`
     - Hoặc cho một giải pháp hợp lý (document rõ trong comment code).

   ### Yêu cầu nội dung `source.zip`
   - `source.zip` phải là file zip chứa mã nguồn module tại `Dir`:
     - Không include cache file thừa.
     - Bỏ qua thư mục `.git`.
   - Có thể dùng `archive/zip` để tự zip.

4. **Chạy song song (parallel)**
   - Chương trình phải hỗ trợ chạy song song khi pack các module.
   - Cần có:
     - Flag `--concurrency` (hoặc `-j`) để cấu hình số lượng worker (mặc định, ví dụ: 4).
     - Implement 1 worker pool:
       - Một goroutine producer push danh sách module cần pack.
       - N worker goroutine tiêu thụ và pack.
   - Phải đảm bảo thread-safe khi:
     - Ghi log.
     - Tạo thư mục dưới `storage_root`.
   - Phải cẩn thận khi skip module đã tồn tại (tránh race condition nếu 2 worker cùng xử lý 1 module).

5. **Idempotent + skip module đã có**
   - Trước khi pack module `Path@Version`, chương trình phải kiểm tra:
     - Nếu `source.zip` đã tồn tại trong `${storage_root}/${Path}/${Version}/source.zip`:
       - In log: đã tồn tại, skip.
       - Không thực hiện pack lại.
   - Điều này cho phép:
     - Chạy lại chương trình nhiều lần mà không mất thời gian pack lại những module đã có.
     - Dễ update incremental.

6. **CLI interface & flags**
   - Chương trình chính (ví dụ `cmd/athens-prefill/main.go`) cho phép các flag:
     - `--modules, -m` : đường dẫn file danh sách module (bắt buộc).
     - `--storage-root, -s` : đường dẫn Athens disk storage root (mặc định đọc từ env `ATHENS_DISK_STORAGE_ROOT`, nếu không có thì báo lỗi).
     - `--work-dir, -w` : thư mục làm việc tạm (temp dir, có default).
     - `--concurrency, -j` : số worker (default 4).
     - `--log-level` : `debug`, `info`, `warn`, `error` (optional).
   - Có thể dùng thư viện flag chuẩn hoặc một thư viện CLI như `spf13/cobra`. Tôi ưu tiên **cobra** cho CLI đẹp, có subcommand, help auto.

7. **Kiến trúc project**

   Hãy tổ chức project theo dạng clean & dễ đọc, ví dụ:

```
├── cmd
│ └── athens-prefill
│ └── main.go
├── internal
│ ├── cli // parse flags, setup config
│ ├── gomod // wrapper gọi 'go' / xử lý go modules
│ ├── resolver // build danh sách module@version@dir từ input + dependencies
│ ├── packer // logic pack module ra Athens format
│ ├── worker // worker pool / concurrency
│ └── log // setup logger
├── go.mod
├── go.sum
└── README.md
```

- Tránh đặt toàn bộ logic trong `main.go`.
- Tách riêng:
  - `resolver` chịu trách nhiệm chạy `go` commands, parse JSON `go list -m -json all` thành struct.
  - `packer` chỉ nhận input là struct `Module{Path, Version, Dir}` và `storageRoot`, sau đó pack ra file.
  - `worker` xử lý concurrent execution.

8. **Logging & UX**
- Sử dụng log có level, ví dụ dùng `log/slog` (Go 1.21+) hoặc một lib phổ biến (zap, zerolog).
- Log cần:
  - Thông tin module đang xử lý: `path`, `version`.
  - Thông báo khi skip module đã tồn tại.
  - Thông báo lỗi khi pack hoặc khi gọi lệnh `go` fail.
- Ở chế độ bình thường (`info`), log ngắn gọn; ở `debug` thì có thể in chi tiết hơn (stdout).

9. **Error handling**
- Nếu một module bị lỗi pack (ví dụ thiếu `Dir`, folder không tồn tại, zip lỗi):
  - Log error.
  - Không làm crash toàn chương trình (trừ khi lỗi hệ thống).
  - Cuối chương trình in ra summary:
    - Tổng số module.
    - Số module pack thành công.
    - Số module lỗi (liệt kê ngắn gọn).

10. **Tests**
 - Viết một số test cơ bản cho:
   - Hàm parse file `modules.txt` thành danh sách specs (path + optional version).
   - Hàm kiểm tra emtpy/skip logic (nếu file `source.zip` đã tồn tại).
   - Hàm build đường dẫn target trong storage.
 - Không cần test integration đầy đủ (vì phụ thuộc vào `go` binary), nhưng nên có test unit.

11. **README.md**
 Hãy tạo file `README.md` mô tả:

 - Mục đích của tool.
 - Cách build:
   ```bash
   go build ./cmd/athens-prefill
   ```
 - Cách sử dụng ví dụ:
   ```bash
   export ATHENS_DISK_STORAGE_ROOT=/data/athens-storage
   ./athens-prefill \
     --modules ./modules.txt \
     --storage-root /data/athens-storage \
     --concurrency 8
   ```
 - Cách copy storage sang máy offline và cấu hình Athens sử dụng `StorageType = "disk"`.

## Yêu cầu output

Hãy trả về:

1. File `go.mod` đầy đủ.
2. Cấu trúc thư mục project (dưới dạng cây).
3. Toàn bộ code Golang cần thiết (các file `.go` chính), đủ để:
- Tôi chỉ cần `git init`, `go mod tidy`, `go build ./cmd/athens-prefill` là compile được.
4. File `README.md` với nội dung rõ ràng.
5. Nếu bạn cần giả lập một số phần (ví dụ không thể gọi thực sự `go list`), hãy comment rõ trong code và vẫn cố gắng đưa ra phiên bản gần với thực tế nhất có thể.

Hãy bắt đầu từ phần `go.mod`, sau đó đến `tree` cấu trúc thư mục, rồi lần lượt các file `.go` và cuối cùng là `README.md`.
