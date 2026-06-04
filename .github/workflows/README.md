# Workflows trong thư mục này (legacy)

`develop.yml` và `staging.yaml` dùng **Harbor + Kubernetes** (Enspire).

Repo monorepo **CleverSchool** dùng workflow ở root:

`.github/workflows/deploy-api.yml` → VPS Docker (`api-dev.viethung.uk`, `api-staging.viethung.uk`).

Không cần copy file từ đây lên root trừ khi vẫn deploy K8s.
