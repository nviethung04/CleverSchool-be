class FileUploader {
  constructor() {
    this.files = [];
    this.uploads = [];
    this.params = this.getUrlParams();
    this.initializeElements();
    this.setupDimensions();
    this.bindEvents();
  }

  initializeElements() {
    this.uploadArea = document.getElementById("uploadArea");
    this.fileInput = document.getElementById("fileInput");
    this.selectBtn = document.getElementById("selectBtn");
    this.uploadBtn = document.getElementById("uploadBtn");
    this.clearBtn = document.getElementById("clearBtn");
    this.fileList = document.getElementById("fileList");
    this.filesContainer = document.getElementById("files");
    this.uploadProgress = document.getElementById("uploadProgress");
    this.progressContainer = document.getElementById("progressContainer");

    // Container element for sizing
    this.container = document.querySelector(".container");
    this.paramInfo = document.getElementById("paramInfo");
  }

  bindEvents() {
    // File selection
    this.selectBtn.addEventListener("click", () => this.fileInput.click());
    this.fileInput.addEventListener("change", (e) =>
      this.handleFileSelection(e)
    );

    // Drag and drop
    this.uploadArea.addEventListener("dragover", (e) => this.handleDragOver(e));
    this.uploadArea.addEventListener("dragleave", (e) =>
      this.handleDragLeave(e)
    );
    this.uploadArea.addEventListener("drop", (e) => this.handleDrop(e));
    this.uploadArea.addEventListener("click", () => this.fileInput.click());

    // Upload actions
    this.uploadBtn.addEventListener("click", () => this.startUpload());
    this.clearBtn.addEventListener("click", () => this.clearAllFiles());

    // Prevent default drag behaviors
    ["dragenter", "dragover", "dragleave", "drop"].forEach((eventName) => {
      document.body.addEventListener(eventName, this.preventDefaults, false);
    });
  }

  preventDefaults(e) {
    e.preventDefault();
    e.stopPropagation();
  }

  getUrlParams() {
    const urlParams = new URLSearchParams(window.location.search);
    return {
      width: urlParams.get("width") || "800",
      height: urlParams.get("height") || "600",
      path: urlParams.get("path") || "uploads",
    };
  }

  setupDimensions() {
    // Apply dimensions to container
    if (this.container) {
      this.container.style.width = this.params.width + "px";
      this.container.style.height = this.params.height + "px";
      this.container.style.overflow = "auto";
      this.container.style.display = "block";
    }

    // Set document title with path info
    document.title = `Upload File - ${this.params.path} - Clever School Clever School`;

    // Update param info display
    if (this.paramInfo) {
      this.paramInfo.innerHTML = `
                📏 Kích thước: ${this.params.width}px × ${this.params.height}px |
                📁 Đường dẫn: /${this.params.path}
            `;
    }
  }

  handleDragOver(e) {
    this.preventDefaults(e);
    this.uploadArea.classList.add("dragover");
  }

  handleDragLeave(e) {
    this.preventDefaults(e);
    this.uploadArea.classList.remove("dragover");
  }

  handleDrop(e) {
    this.preventDefaults(e);
    this.uploadArea.classList.remove("dragover");

    const files = Array.from(e.dataTransfer.files);
    this.addFiles(files);
  }

  handleFileSelection(e) {
    const files = Array.from(e.target.files);
    this.addFiles(files);
    this.fileInput.value = ""; // Reset input
  }

  addFiles(newFiles) {
    newFiles.forEach((file) => {
      if (
        !this.files.find((f) => f.name === file.name && f.size === file.size)
      ) {
        this.files.push(file);
      }
    });

    this.updateFileList();
    this.updateButtons();
  }

  removeFile(index) {
    this.files.splice(index, 1);
    this.updateFileList();
    this.updateButtons();
  }

  updateFileList() {
    if (this.files.length === 0) {
      this.fileList.style.display = "none";
      return;
    }

    this.fileList.style.display = "block";
    this.filesContainer.innerHTML = "";

    this.files.forEach((file, index) => {
      const fileItem = this.createFileItem(file, index);
      this.filesContainer.appendChild(fileItem);
    });
  }

  createFileItem(file, index) {
    const fileItem = document.createElement("div");
    fileItem.className = "file-item";

    const fileIcon = this.getFileIcon(file);
    const fileSize = this.formatFileSize(file.size);

    fileItem.innerHTML = `
            <div class="file-info">
                <div class="file-icon ${fileIcon.class}">
                    <i class="${fileIcon.icon}"></i>
                </div>
                <div class="file-details">
                    <h4>${file.name}</h4>
                    <p>${file.type || "Unknown type"}</p>
                </div>
                <div class="file-size">${fileSize}</div>
            </div>
            <div class="file-actions">
                <button class="remove-btn" title="Xóa file">
                    <i class="fas fa-times"></i>
                </button>
            </div>
        `;

    // Add remove event
    const removeBtn = fileItem.querySelector(".remove-btn");
    removeBtn.addEventListener("click", () => this.removeFile(index));

    return fileItem;
  }

  getFileIcon(file) {
    const type = file.type;
    const name = file.name.toLowerCase();

    if (type.startsWith("image/")) {
      return { icon: "fas fa-image", class: "image" };
    } else if (type.startsWith("video/")) {
      return { icon: "fas fa-video", class: "video" };
    } else if (type.startsWith("audio/")) {
      return { icon: "fas fa-music", class: "audio" };
    } else if (type.includes("pdf")) {
      return { icon: "fas fa-file-pdf", class: "pdf" };
    } else if (type.includes("word") || type.includes("document")) {
      return { icon: "fas fa-file-word", class: "document" };
    } else if (type.includes("excel") || type.includes("spreadsheet")) {
      return { icon: "fas fa-file-excel", class: "document" };
    } else if (type.includes("powerpoint") || type.includes("presentation")) {
      return { icon: "fas fa-file-powerpoint", class: "document" };
    } else if (
      name.endsWith(".zip") ||
      name.endsWith(".rar") ||
      name.endsWith(".7z")
    ) {
      return { icon: "fas fa-file-archive", class: "archive" };
    } else {
      return { icon: "fas fa-file", class: "other" };
    }
  }

  formatFileSize(bytes) {
    if (bytes === 0) return "0 Bytes";

    const k = 1024;
    const sizes = ["Bytes", "KB", "MB", "GB", "TB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));

    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
  }

  updateButtons() {
    const hasFiles = this.files.length > 0;
    this.uploadBtn.disabled = !hasFiles;
    this.clearBtn.disabled = !hasFiles;
  }

  clearAllFiles() {
    this.files = [];
    this.updateFileList();
    this.updateButtons();
    this.hideProgress();
  }

  startUpload() {
    if (this.files.length === 0) return;

    this.showProgress();
    this.uploadBtn.disabled = true;

    this.files.forEach((file, index) => {
      this.uploadFile(file, index);
    });
  }

  uploadFile(file, index) {
    const upload = {
      file: file,
      index: index,
      status: "uploading",
      progress: 0,
      element: null,
      options: this.getUploadOptions(),
    };

    this.uploads.push(upload);
    this.createProgressItem(upload);

    // Resolve API base to avoid mixed content when page is HTTPS but API_DOMAIN is HTTP on same host
    const apiBase = (() => {
      try {
        const page = window.location;
        const configured = window.API_DOMAIN;
        if (!configured) return page.origin;
        const cfg = new URL(configured, page.origin);
        // If page is https and configured is http but same host, upgrade to https
        if (
          page.protocol === "https:" &&
          cfg.protocol === "http:" &&
          cfg.hostname === page.hostname
        ) {
          cfg.protocol = "https:";
        }
        return cfg.origin;
      } catch (e) {
        return window.location.origin;
      }
    })();

    console.log("Uploading to:", apiBase + "/api/tus-uploads/");

    // Create TUS upload with options
    const tusUpload = new tus.Upload(file, {
      endpoint: apiBase + "/api/tus-uploads/",
      retryDelays: [0, 1000, 3000, 5000, 10000, 30000], // Thêm retry delays dài hơn
      chunkSize: 8 * 1024 * 1024, // 8MB chunks - phải khớp với server
      removeFingerprintOnSuccess: true, // Xóa fingerprint sau khi thành công
      uploadLengthDeferred: false, // Không sử dụng deferred length
      // Thêm cấu hình để tránh upload bị reset
      onBeforeRequest: (req) => {
        // Log request để debug
        console.log(`Uploading chunk: ${req.getURL()}`);
        // Thêm timeout cho mỗi request
        req.setHeader('X-Request-Timeout', '300000'); // 5 phút
      },
      metadata: {
        filename: file.name,
        filetype: file.type,
        filesize: file.size.toString(),
        width: this.getUploadOptions().width,
        height: this.getUploadOptions().height,
        path: this.getUploadOptions().path,
      },
      onError: (error) => {
        console.error("Upload failed:", error);
        // Thử resume upload nếu có thể
        if (error.originalRequest && error.originalRequest.getUnderlyingObject) {
          const underlying = error.originalRequest.getUnderlyingObject();
          if (underlying && underlying.status === 0) {
            console.log("Network error, attempting to resume...");
            // Thử resume sau 5 giây
            setTimeout(() => {
              tusUpload.start();
            }, 5000);
            return;
          }
        }
        this.updateUploadStatus(upload, "error", "Upload failed");
      },
      onProgress: (bytesUploaded, bytesTotal) => {
        const progress = (bytesUploaded / bytesTotal) * 100;
        this.updateUploadProgress(upload, progress);
      },
      onSuccess: () => {
        this.updateUploadStatus(upload, "success", "Upload completed");
        // this.processUploadSuccess(upload);
        this.checkAllUploadsComplete();
      },
      onChunkComplete: (chunkSize, bytesAccepted, bytesTotal) => {
        console.log(`Chunk completed: ${bytesAccepted}/${bytesTotal} bytes`);
      },
    });

    // Start upload
    tusUpload.start();
  }

  createProgressItem(upload) {
    const progressItem = document.createElement("div");
    progressItem.className = "progress-item";
    progressItem.innerHTML = `
            <div class="progress-header">
                <div class="progress-filename">${upload.file.name}</div>
                <div class="progress-status">Đang upload...</div>
            </div>
            <div class="progress-bar">
                <div class="progress-fill" style="width: 0%"></div>
            </div>
            <div class="progress-details">
                <span>0%</span>
                <span>${this.formatFileSize(upload.file.size)}</span>
            </div>
        `;

    this.progressContainer.appendChild(progressItem);
    upload.element = progressItem;
  }

  updateUploadProgress(upload, progress) {
    upload.progress = progress;

    if (upload.element) {
      const progressFill = upload.element.querySelector(".progress-fill");
      const progressPercent = upload.element.querySelector(
        ".progress-details span:first-child"
      );

      if (progressFill) progressFill.style.width = progress + "%";
      if (progressPercent)
        progressPercent.textContent = Math.round(progress) + "%";
    }
  }

  updateUploadStatus(upload, status, message) {
    upload.status = status;

    if (upload.element) {
      const progressStatus = upload.element.querySelector(".progress-status");
      if (progressStatus) progressStatus.textContent = message;

      upload.element.classList.add(status);
    }
  }

  checkAllUploadsComplete() {
    const allComplete = this.uploads.every(
      (upload) => upload.status === "success" || upload.status === "error"
    );

    if (allComplete) {
      this.uploadBtn.disabled = false;
      this.showCompletionMessage();
    }
  }

  showCompletionMessage() {
    const successCount = this.uploads.filter(
      (u) => u.status === "success"
    ).length;
    const errorCount = this.uploads.filter((u) => u.status === "error").length;

    let message = "";
    if (successCount > 0 && errorCount === 0) {
      message = `✅ Tất cả ${successCount} file đã được upload thành công!`;
    } else if (successCount > 0 && errorCount > 0) {
      message = `⚠️ ${successCount} file thành công, ${errorCount} file thất bại.`;
    } else {
      message = `❌ Tất cả file upload đều thất bại.`;
    }

    // Show notification
    this.showNotification(message, successCount > 0 ? "success" : "error");
  }

  showNotification(message, type) {
    const notification = document.createElement("div");
    notification.className = `notification ${type}`;
    notification.textContent = message;

    // Add styles
    notification.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            padding: 15px 20px;
            border-radius: 8px;
            color: white;
            font-weight: 600;
            z-index: 1000;
            animation: slideIn 0.3s ease;
            background: ${type === "success" ? "#28a745" : "#dc3545"};
        `;

    document.body.appendChild(notification);

    // Remove after 5 seconds
    setTimeout(() => {
      notification.remove();
    }, 5000);
  }

  showProgress() {
    this.uploadProgress.style.display = "block";
  }

  hideProgress() {
    this.uploadProgress.style.display = "none";
    this.progressContainer.innerHTML = "";
    this.uploads = [];
  }

  getUploadOptions() {
    return {
      width: this.params.width,
      height: this.params.height,
      path: this.params.path,
    };
  }

  processUploadSuccess(upload) {
    const options = upload.options;
    this.showNotification(
      `✅ File ${upload.file.name} đã được upload thành công!`,
      "success"
    );

    // Show storage path info
    setTimeout(() => {
      this.showNotification(
        `📁 File đã được lưu vào thư mục: ${options.path}`,
        "info"
      );
    }, 1000);
  }
}

// Add CSS animation for notification
const style = document.createElement("style");
style.textContent = `
    @keyframes slideIn {
        from {
            transform: translateX(100%);
            opacity: 0;
        }
        to {
            transform: translateX(0);
            opacity: 1;
        }
    }
`;
document.head.appendChild(style);

// Initialize when DOM is loaded
document.addEventListener("DOMContentLoaded", () => {
  new FileUploader();
});
