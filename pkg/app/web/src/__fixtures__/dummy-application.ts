import { ApplicationKind } from "pipe/pkg/app/web/model/common_pb";
import {
  Application,
  ApplicationDeploymentReference,
  ApplicationSyncState,
  ApplicationSyncStatus,
} from "~/modules/applications";
import { createGitPathFromObject } from "./common";
import { dummyEnv } from "./dummy-environment";
import { dummyPiped } from "./dummy-piped";
import { dummyRepo } from "./dummy-repo";
import { createTriggerFromObject, dummyTrigger } from "./dummy-trigger";
import { createRandTimes, randomUUID } from "./utils";

export const dummyApplicationSyncState: ApplicationSyncState.AsObject = {
  headDeploymentId: "deployment-1",
  reason: "",
  shortReason: "",
  status: ApplicationSyncStatus.SYNCED,
  timestamp: 0,
};

const [createdAt, startedAt, updatedAt] = createRandTimes(3);

export const dummyApplication: Application.AsObject = {
  id: randomUUID(),
  cloudProvider: "kubernetes-default",
  disabled: false,
  envId: dummyEnv.id,
  gitPath: {
    configPath: "",
    configFilename: "",
    path: "dir/dir1",
    url: "",
    repo: dummyRepo,
  },
  kind: ApplicationKind.KUBERNETES,
  name: "DemoApp",
  pipedId: dummyPiped.id,
  projectId: "project-1",
  description: "",
  labelsMap: [],
  mostRecentlySuccessfulDeployment: {
    deploymentId: "deployment-1",
    completedAt: 0,
    summary: "",
    startedAt: startedAt.unix(),
    version: "v1",
    trigger: dummyTrigger,
  },
  mostRecentlyTriggeredDeployment: {
    deploymentId: "deployment-1",
    completedAt: 0,
    summary: "summary",
    startedAt: startedAt.unix(),
    version: "v1",
    trigger: dummyTrigger,
  },
  syncState: dummyApplicationSyncState,
  updatedAt: updatedAt.unix(),
  deletedAt: 0,
  createdAt: createdAt.unix(),
  deleted: false,
  deploying: false,
};

export const dummyApps: Record<ApplicationKind, Application.AsObject> = {
  [ApplicationKind.KUBERNETES]: dummyApplication,
  [ApplicationKind.TERRAFORM]: {
    ...dummyApplication,
    id: randomUUID(),
    name: "Terraform App",
    kind: ApplicationKind.TERRAFORM,
    cloudProvider: "terraform-default",
  },
  [ApplicationKind.LAMBDA]: {
    ...dummyApplication,
    id: randomUUID(),
    name: "Lambda App",
    kind: ApplicationKind.LAMBDA,
    cloudProvider: "lambda-default",
  },
  [ApplicationKind.CLOUDRUN]: {
    ...dummyApplication,
    id: randomUUID(),
    name: "CloudRun App",
    kind: ApplicationKind.CLOUDRUN,
    cloudProvider: "cloud-run-default",
  },
  [ApplicationKind.ECS]: {
    ...dummyApplication,
    id: randomUUID(),
    name: "ECS App",
    kind: ApplicationKind.ECS,
    cloudProvider: "ecs-default",
  },
};

function createAppSyncStateFromObject(
  o: ApplicationSyncState.AsObject
): ApplicationSyncState {
  const state = new ApplicationSyncState();
  state.setHeadDeploymentId(o.headDeploymentId);
  state.setReason(o.reason);
  state.setShortReason(o.shortReason);
  state.setStatus(o.status);
  state.setTimestamp(o.timestamp);
  return state;
}

function createAppDeploymentReferenceFromObject(
  o: ApplicationDeploymentReference.AsObject
): ApplicationDeploymentReference {
  const ref = new ApplicationDeploymentReference();
  ref.setDeploymentId(o.deploymentId);
  ref.setSummary(o.summary);
  ref.setVersion(o.version);
  ref.setStartedAt(o.startedAt);
  ref.setCompletedAt(o.completedAt);
  if (o.trigger) {
    ref.setTrigger(createTriggerFromObject(o.trigger));
  }
  return ref;
}

export function createApplicationFromObject(
  o: Application.AsObject
): Application {
  const app = new Application();
  app.setId(o.id);
  app.setCloudProvider(o.cloudProvider);
  app.setDisabled(o.disabled);
  app.setEnvId(o.envId);
  app.setKind(o.kind);
  app.setName(o.name);
  app.setPipedId(o.pipedId);
  app.setProjectId(o.projectId);
  app.setCreatedAt(o.createdAt);
  app.setUpdatedAt(o.updatedAt);
  app.setDeletedAt(o.deletedAt);
  app.setDeleted(o.deleted);
  app.setDeploying(o.deploying);
  if (o.syncState) {
    app.setSyncState(createAppSyncStateFromObject(o.syncState));
  }
  if (o.gitPath) {
    app.setGitPath(createGitPathFromObject(o.gitPath));
  }
  if (o.mostRecentlySuccessfulDeployment) {
    app.setMostRecentlySuccessfulDeployment(
      createAppDeploymentReferenceFromObject(o.mostRecentlySuccessfulDeployment)
    );
  }
  if (o.mostRecentlyTriggeredDeployment) {
    app.setMostRecentlyTriggeredDeployment(
      createAppDeploymentReferenceFromObject(o.mostRecentlyTriggeredDeployment)
    );
  }
  return app;
}
