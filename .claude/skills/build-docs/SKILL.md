---
name: build-docs
description: >
  Xây dựng bộ tài liệu tiếng Việt về một thành phần/công nghệ cụ thể.
  Tài liệu giải thích khái niệm, công dụng, cách hoạt động, manifest mẫu, và các lưu ý thực tế.
  TRIGGER khi user yêu cầu viết docs, tài liệu, giải thích một thành phần hạ tầng/công nghệ.
argument-hint: "[tên-thành-phần] [thư-mục-đích (tuỳ chọn)]"
arguments: [component, target_dir]
user-invocable: true
allowed-tools: Read Grep Glob Bash WebSearch WebFetch Write Edit Agent
effort: high
---

# Build Docs — Tạo tài liệu kỹ thuật tiếng Việt

## Thành phần cần viết docs

**Component:** `$component`
**Target directory:** `$target_dir` (nếu không truyền, tự suy từ cấu trúc repo)

---

## Quy trình thực hiện

### Bước 1 — Nghiên cứu từ nguồn đáng tin cậy

Trước khi viết bất kỳ dòng nào, BẮT BUỘC phải research kỹ về `$component`:

1. **Docs chính thức** — Tìm và đọc documentation chính thức của `$component` (kubernetes.io, docs của project, RFC, spec). Đây là nguồn ưu tiên cao nhất.
2. **Blog kỹ thuật uy tín** — Tham khảo thêm từ các nguồn như: CNCF blog, official project blog, AWS/GCP/Azure docs, DigitalOcean tutorials, Red Hat developer blog.
3. **Cộng đồng** — GitHub issues/discussions, Stack Overflow answers có vote cao, các bài phân tích kỹ thuật chuyên sâu.

**Dùng `WebSearch` để tìm kiếm** với các query:
- `"$component" official documentation`
- `"$component" architecture how it works`
- `"$component" best practices production`
- `"$component" common pitfalls`

**Dùng `WebFetch` để đọc nội dung** từ các URL tìm được.

> QUAN TRỌNG: Không bịa thông tin. Mọi khái niệm kỹ thuật phải có cơ sở từ nguồn đáng tin cậy. Nếu không chắc chắn, ghi rõ hoặc bỏ qua.

---

### Bước 2 — Xác định style từ docs hiện có trong repo

Đọc các README.md hiện có trong repo (đặc biệt trong `tech/k8s/`) để đảm bảo style nhất quán.

Dùng lệnh sau để tìm docs mẫu:

```!
find /d/nampham2003/challenges/tech -name "README.md" -type f | head -10
```

Đọc ít nhất 2 file README.md gần nhất với chủ đề đang viết để nắm style.

---

### Bước 3 — Viết docs theo template

Viết file `README.md` bằng **tiếng Việt**, tuân thủ CHÍNH XÁC template sau:

````markdown
# [Tên thành phần] trong [Hệ sinh thái]

## [Tên thành phần] là gì?

**[Tên]** là [loại + định nghĩa ngắn gọn]. [Hành vi/đảm bảo chính]. [So sánh với khái niệm liên quan nếu cần].

---

## Vấn đề mà [Tên] giải quyết

[Liệt kê các vấn đề cụ thể]:

- **Vấn đề 1:** Mô tả
- **Vấn đề 2:** Mô tả
- **Vấn đề 3:** Mô tả

[Tên] giải quyết bằng cách [cơ chế cốt lõi].

---

## Cách hoạt động

[Giải thích chi tiết cơ chế bên trong]

### [Khía cạnh 1]

[Chi tiết]

### [Khía cạnh 2]

[Chi tiết]

---

## Cấu trúc Manifest

```yaml
apiVersion: ...
kind: ...
metadata:
  name: example
spec:
  field1: value      # giải thích ngắn — tập trung vào "tại sao"
  field2: value      # comment cô đọng, không giải thích definition
```

---

## Các trường quan trọng

### `fieldName`

[Mô tả chức năng, tác dụng]

[Bảng so sánh các tuỳ chọn nếu có]

| Tuỳ chọn | Hành vi |
|-----------|---------|
| option1   | ...     |
| option2   | ...     |

---

## [Các mục bổ sung tuỳ theo chủ đề]

[Nội dung với bảng, ví dụ, so sánh]

---

## Lưu ý thực tế (Common Pitfalls)

**Lưu ý 1 — [Mô tả ngắn]:** [Giải thích đầy đủ và cách tránh]

**Lưu ý 2 — [Mô tả ngắn]:** [Giải thích đầy đủ và cách tránh]

---

## Tham khảo

- [Tên - Official Docs](https://...)
- [Thành phần liên quan](../path/README.md)
- [Nguồn bổ sung](https://...)
````

---

### Bước 4 — Kiểm tra chất lượng

Trước khi hoàn thành, tự kiểm tra:

- [ ] Đã research từ docs chính thức, KHÔNG bịa thông tin
- [ ] Viết bằng tiếng Việt, thuật ngữ kỹ thuật giữ nguyên tiếng Anh (VD: Pod, Service, Controller)
- [ ] Theo đúng template: What → Problem → How → Manifest → Key Fields → Pitfalls → References
- [ ] YAML manifest có comment cô đọng 1-2 dòng, giải thích decision không giải thích definition
- [ ] Có bảng so sánh khi có nhiều tuỳ chọn/tradeoff
- [ ] Dùng `---` ngăn cách giữa các section chính
- [ ] **Bold** cho thuật ngữ quan trọng lần đầu xuất hiện
- [ ] Backtick cho resource names, commands, paths
- [ ] Section "Tham khảo" có link docs chính thức + cross-link tới docs liên quan trong repo
- [ ] Không thừa — mỗi khái niệm giải thích một lần, cross-reference thay vì lặp
- [ ] Ví dụ YAML production-ready, không đơn giản hoá quá mức

---

## Quy tắc style bắt buộc

1. **Theory-first**: Giải thích khái niệm trước, ví dụ sau
2. **Problem-before-solution**: Luôn nêu vấn đề trước khi đưa giải pháp
3. **Cô đọng**: Chi tiết nhưng không dài dòng, không prose thừa
4. **Production-focused**: Ví dụ và cảnh báo hướng tới môi trường production thực tế
5. **Tiếng Việt tự nhiên**: Viết tiếng Việt trôi chảy, thuật ngữ kỹ thuật giữ nguyên English
6. **Comment style**: Cô đọng 1-2 dòng, giải thích decision không giải thích definition
7. **Luôn dùng PostgreSQL** thay MySQL trong các ví dụ database
8. **Cross-linking**: Dùng relative path link tới docs liên quan: `[Tên](../path/README.md)`
9. **Horizontal dividers**: Dùng `---` giữa các section H2
10. **Em-dash**: Dùng `—` cho các ghi chú phụ inline
