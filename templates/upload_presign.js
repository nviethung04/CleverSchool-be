const fileInput = document.getElementById("fileInput");
const selectBtn = document.getElementById("selectBtn");
const uploadBtn = document.getElementById("uploadBtn");
const clearBtn = document.getElementById("clearBtn");
const fileList = document.getElementById("fileList");
const filesDiv = document.getElementById("files");
const uploadProgress = document.getElementById("uploadProgress");
const progressContainer = document.getElementById("progressContainer");
const uploadArea = document.getElementById("uploadArea");

let selectedFiles = [];

selectBtn.addEventListener("click", () => fileInput.click());
fileInput.addEventListener("change", (e) => handleFiles(e.target.files));
clearBtn.addEventListener("click", clearFiles);

uploadArea.addEventListener("dragover", (e) => {
  e.preventDefault();
  uploadArea.classList.add("dragover");
});
uploadArea.addEventListener("dragleave", () =>
  uploadArea.classList.remove("dragover")
);
uploadArea.addEventListener("drop", (e) => {
  e.preventDefault();
  uploadArea.classList.remove("dragover");
  handleFiles(e.dataTransfer.files);
});

function handleFiles(files) {
  for (let f of files) selectedFiles.push(f);
  renderFileList();
}

function renderFileList() {
  filesDiv.innerHTML = "";
  if (selectedFiles.length === 0) {
    fileList.style.display = "none";
    uploadBtn.disabled = true;
    clearBtn.disabled = true;
    return;
  }
  fileList.style.display = "block";
  uploadBtn.disabled = false;
  clearBtn.disabled = false;

  selectedFiles.forEach((file, idx) => {
    const div = document.createElement("div");
    div.className = "file-item";
    div.innerHTML = `<i class="fas fa-file"></i> ${file.name} (${(
      file.size /
      1024 /
      1024
    ).toFixed(2)} MB)`;
    filesDiv.appendChild(div);
  });
}

function clearFiles() {
  selectedFiles = [];
  renderFileList();
  progressContainer.innerHTML = "";
  uploadProgress.style.display = "none";
}

uploadBtn.addEventListener("click", async () => {
  if (selectedFiles.length === 0) return;
  uploadProgress.style.display = "block";
  progressContainer.innerHTML = "";

  for (let file of selectedFiles) {
    await uploadFile(file);
  }
});

async function uploadFile(file) {
  // 🟣 1. Tạo khối hiển thị tiến trình upload
  const progressItem = document.createElement("div");
  progressItem.className = "progress-item";
  progressItem.innerHTML = `
    <div class="progress-header">
      <span class="progress-filename">${file.name}</span>
      <span class="progress-status">Đang tải...</span>
    </div>
    <div class="progress-bar">
      <div class="progress-fill" style="width:0%"></div>
    </div>
    <div class="progress-details">
      <span class="progress-percent">0%</span>
    </div>
  `;
  progressContainer.appendChild(progressItem);

  // Gán biến để cập nhật UI
  const fill = progressItem.querySelector(".progress-fill");
  const percentText = progressItem.querySelector(".progress-percent");
  const statusText = progressItem.querySelector(".progress-status");

  try {
    // 🟣 2. Lấy presign URL từ backend
    const urlParams = new URLSearchParams(window.location.search);
    const folder = urlParams.get("path") || "";

    const presignRes = await fetch(`${window.API_DOMAIN}/api/upload/presign`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        filename: file.name,
        content_type: file.type,
        folder: folder,
      }),
    });

    const { data } = await presignRes.json();
    if (!data?.upload_url) throw new Error("Không nhận được upload URL");

    // 🟣 3. Upload lên S3
    await new Promise((resolve, reject) => {
      const xhr = new XMLHttpRequest();
      xhr.open("PUT", data.upload_url, true);
      xhr.setRequestHeader(
        "Content-Type",
        file.type || "application/octet-stream"
      );

      xhr.upload.addEventListener("progress", (e) => {
        if (e.lengthComputable) {
          const percent = (e.loaded / e.total) * 100;
          fill.style.width = percent.toFixed(2) + "%";
          percentText.textContent = percent.toFixed(0) + "%";
        }
      });

      xhr.onload = () => {
        if (xhr.status >= 200 && xhr.status < 300) {
          fill.style.width = "100%";
          statusText.textContent = "Đang xử lý... ⏳";
          fill.style.background = "#00bcd4"; // xanh lam nhạt
          progressItem.classList.add("processing");
          resolve();
        } else {
          progressItem.classList.add("error");
          fill.style.background = "red";
          statusText.textContent = "Lỗi tải lên ❌";
          reject(new Error(`Upload failed: ${xhr.status}`));
        }
      };

      xhr.onerror = () => {
        progressItem.classList.add("error");
        fill.style.background = "red";
        statusText.textContent = "Mất kết nối ⚠️";
        reject(new Error("Network error"));
      };

      xhr.send(file);
    });

    // 🟠 4. Gọi API /complete (có thể tốn thời gian)
    const completeRes = await fetch(
      `${window.API_DOMAIN}/api/upload/complete`,
      {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          filename: file.name,
          key: data.key,
          url: data.object_url,
          size: file.size,
          content_type: file.type,
          user_id: 123,
        }),
      }
    );

    if (!completeRes.ok)
      throw new Error(`Complete failed: ${completeRes.status}`);

    // Nếu backend trả về thêm data
    const completeData = await completeRes.json();
    const finalUrl = completeData?.data?.final_url || data.object_url;

    // ✅ 5. Hoàn tất
    statusText.textContent = "Thành công 🎉";
    progressItem.classList.remove("processing");
    progressItem.classList.add("success");
    fill.style.background = "#4caf50"; // xanh lá
    showUploadedLink(file.name, finalUrl);
  } catch (err) {
    progressItem.classList.add("error");
    fill.style.background = "red";
    statusText.textContent = "Thất bại ❌";
    console.error("Upload error:", err);
  }
}

function showUploadedLink(filename, url) {
  let linkContainer = document.getElementById("uploadedLinks");
  linkContainer.style.display = "block";

  const linkItem = document.createElement("div");
  linkItem.className = "uploaded-link";
  linkItem.innerHTML = `
    <div class="link-item">
      <span><i class="fas fa-link"></i> ${filename}</span>
      <input type="text" readonly value="${url}" class="link-input" onclick="this.select()">
      <button class="copy-btn" onclick="navigator.clipboard.writeText('${url}')">
        <i class="fas fa-copy"></i> Copy
      </button>
    </div>
  `;
  linkContainer.appendChild(linkItem);
}
