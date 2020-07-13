import { apiClient, apiRequest } from "./client";
import {
  GetDeploymentRequest,
  GetDeploymentResponse,
  ListDeploymentsRequest,
  ListDeploymentsResponse,
  CancelDeploymentRequest,
  CancelDeploymentResponse,
  ApproveStageRequest,
  ApproveStageResponse,
} from "pipe/pkg/app/web/api_client/service_pb";

export const getDeployment = ({
  deploymentId,
}: GetDeploymentRequest.AsObject): Promise<GetDeploymentResponse.AsObject> => {
  const req = new GetDeploymentRequest();
  req.setDeploymentId(deploymentId);
  return apiRequest(req, apiClient.getDeployment);
};

export const getDeployments = ({
  options,
}: ListDeploymentsRequest.AsObject): Promise<
  ListDeploymentsResponse.AsObject
> => {
  const req = new ListDeploymentsRequest();
  if (options) {
    const opts = new ListDeploymentsRequest.Options();
    opts.setEnvIdsList(options.envIdsList);
    opts.setApplicationIdsList(options.applicationIdsList);
    opts.setKindsList(options.kindsList);
    opts.setStatusesList(options.statusesList);
    req.setOptions(opts);
  }
  return apiRequest(req, apiClient.listDeployments);
};

export const cancelDeployment = ({
  deploymentId,
  withoutRollback,
}: CancelDeploymentRequest.AsObject): Promise<
  CancelDeploymentResponse.AsObject
> => {
  const req = new CancelDeploymentRequest();
  req.setDeploymentId(deploymentId);
  req.setWithoutRollback(withoutRollback);
  return apiRequest(req, apiClient.cancelDeployment);
};

export const approveStage = ({
  deploymentId,
  stageId,
}: ApproveStageRequest.AsObject): Promise<ApproveStageResponse.AsObject> => {
  const req = new ApproveStageRequest();
  req.setDeploymentId(deploymentId);
  req.setStageId(stageId);
  return apiRequest(req, apiClient.approveStage);
};
