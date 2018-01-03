// Copyright 2017 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import {actionbarViewName, stateName as chromeStateName} from '../../chrome/state';
import {breadcrumbsConfig} from '../../common/components/breadcrumbs/service';
import {appendDetailParamsToUrl} from '../../common/resource/resourcedetail';
import {stateName as daemonSetList} from '../../daemonset/list/state';

import {stateName as parentState, stateUrl} from '../state';
import {ActionBarController} from './actionbar_controller';
import {DaemonSetDetailController} from './controller';

/**
 * Config state object for the Daemon Set detail view.
 *
 * @type {!ui.router.StateConfig}
 */
export const config = {
  url: appendDetailParamsToUrl(stateUrl),
  parent: parentState,
  resolve: {
    'daemonSetDetailResource': getDaemonSetDetailResource,
    'daemonSetDetail': getDaemonSetDetail,
  },
  data: {
    [breadcrumbsConfig]: {
      'label': '{{$stateParams.objectName}}',
      'parent': daemonSetList,
    },
  },
  views: {
    '': {
      controller: DaemonSetDetailController,
      controllerAs: 'ctrl',
      templateUrl: 'daemonset/detail/detail.html',
    },
    [`${actionbarViewName}@${chromeStateName}`]: {
      controller: ActionBarController,
      controllerAs: '$ctrl',
      templateUrl: 'daemonset/detail/actionbar.html',
    },
  },
};

/**
 * @param {!angular.$resource} $resource
 * @return {!angular.Resource}
 * @ngInject
 */
export function daemonSetPodsResource($resource) {
  return $resource('api/v1/daemonset/:namespace/:name/pod');
}

/**
 * @param {!angular.$resource} $resource
 * @return {!angular.Resource}
 * @ngInject
 */
export function daemonSetServicesResource($resource) {
  return $resource('api/v1/daemonset/:namespace/:name/service');
}

/**
 * @param {!angular.$resource} $resource
 * @return {!angular.Resource}
 * @ngInject
 */
export function daemonSetEventsResource($resource) {
  return $resource('api/v1/daemonset/:namespace/:name/event');
}

/**
 * @param {!./../../common/resource/resourcedetail.StateParams} $stateParams
 * @param {!angular.$resource} $resource
 * @return {!angular.Resource}
 * @ngInject
 */
export function getDaemonSetDetailResource($resource, $stateParams) {
  return $resource(`api/v1/daemonset/${$stateParams.objectNamespace}/${$stateParams.objectName}`);
}

/**
 * @param {!angular.Resource} daemonSetDetailResource
 * @param {!./../../common/resource/resourcedetail.StateParams} $stateParams
 * @param {!./../../common/dataselect/service.DataSelectService} kdDataSelectService
 * @return {!angular.$q.Promise}
 * @ngInject
 */
export function getDaemonSetDetail(daemonSetDetailResource, $stateParams, kdDataSelectService) {
  let query = kdDataSelectService.getDefaultResourceQuery(
      $stateParams.objectNamespace, $stateParams.objectName);
  return daemonSetDetailResource.get(query).$promise;
}
