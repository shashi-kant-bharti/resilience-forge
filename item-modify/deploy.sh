#!/usr/bin/env bash
# deploy.sh — Build, push, and deploy item-modify to IBM Code Engine
#
# Prerequisites (all installed and configured):
#   ibmcloud CLI  +  plugins: code-engine (ce), container-registry (cr)
#   docker
#
# Usage:
#   cd item-modify
#   chmod +x deploy.sh
#   ./deploy.sh
#
# Override defaults by setting env vars before running:
#   ICR_REGION=uk.icr.io  CE_REGION=eu-gb  ./deploy.sh

set -euo pipefail

# ── Configuration ────────────────────────────────────────────────
ICR_REGION="${ICR_REGION:-uk.icr.io}"                        # IBM Container Registry region
ICR_NAMESPACE="${ICR_NAMESPACE:-resilience-forge}"           # CR namespace
IMAGE_NAME="${IMAGE_NAME:-item-modify}"
IMAGE_TAG="${IMAGE_TAG:-latest}"
IMAGE="${ICR_REGION}/${ICR_NAMESPACE}/${IMAGE_NAME}:${IMAGE_TAG}"

CE_REGION="${CE_REGION:-eu-gb}"                              # Code Engine region
CE_PROJECT="${CE_PROJECT:-resilience-forge}"                 # Code Engine project name
CE_APP="${CE_APP:-item-modify}"                              # Code Engine application name
CE_PORT="${CE_PORT:-8080}"

# ── Step 1: Login to IBM Container Registry ──────────────────────
echo "==> Logging in to IBM Container Registry (${ICR_REGION})..."
ibmcloud cr login --client docker

# ── Step 2: Ensure namespace exists ──────────────────────────────
echo "==> Ensuring CR namespace '${ICR_NAMESPACE}' exists..."
ibmcloud cr namespace-add "${ICR_NAMESPACE}" 2>/dev/null || true

# ── Step 3: Build the Docker image ───────────────────────────────
echo "==> Building image ${IMAGE}..."
docker build --platform linux/amd64 -t "${IMAGE}" .

# ── Step 4: Push to IBM Container Registry ───────────────────────
echo "==> Pushing image to ${ICR_REGION}..."
docker push "${IMAGE}"

# ── Step 5: Target the Code Engine project ───────────────────────
echo "==> Targeting Code Engine project '${CE_PROJECT}' in ${CE_REGION}..."
ibmcloud ce project select --name "${CE_PROJECT}" --region "${CE_REGION}" || \
  ibmcloud ce project create --name "${CE_PROJECT}" --region "${CE_REGION}"

# ── Step 6: Deploy (create or update) the application ────────────
echo "==> Deploying '${CE_APP}' to Code Engine..."
if ibmcloud ce app get --name "${CE_APP}" &>/dev/null; then
  ibmcloud ce app update \
    --name        "${CE_APP}" \
    --image       "${IMAGE}" \
    --port        "${CE_PORT}" \
    --env         UI_DIR=/ui \
    --min-scale   1 \
    --max-scale   5
else
  ibmcloud ce app create \
    --name        "${CE_APP}" \
    --image       "${IMAGE}" \
    --port        "${CE_PORT}" \
    --env         UI_DIR=/ui \
    --env         COS_INSTANCE_CRN="${COS_INSTANCE_CRN:?set COS_INSTANCE_CRN}" \
    --env         COS_ENDPOINT="${COS_ENDPOINT:?set COS_ENDPOINT}" \
    --env         COS_BUCKET="${COS_BUCKET:?set COS_BUCKET}" \
    --env         SM_INSTANCE_URL="${SM_INSTANCE_URL:?set SM_INSTANCE_URL}" \
    --env         SM_API_KEY="${SM_API_KEY:?set SM_API_KEY}" \
    --env         SM_SECRET_ID="${SM_SECRET_ID:?set SM_SECRET_ID}" \
    --min-scale   1 \
    --max-scale   5
fi

# ── Step 7: Print the application URL ────────────────────────────
echo ""
echo "==> Deployment complete!"
ibmcloud ce app get --name "${CE_APP}" --output url
