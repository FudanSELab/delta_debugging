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

/**
 * @final
 */
export class PersistentVolumeClaimCardListController {
  /**
   * @ngInject
   * @param {!./../../common/namespace/service.NamespaceService} kdNamespaceService
   */
  constructor(kdNamespaceService) {
    /** @private {!./../../common/namespace/service.NamespaceService} */
    this.kdNamespaceService_ = kdNamespaceService;
    /** @export {!backendApi.PersistentVolumeClaimList} - Initialized from binding. */
    this.persistentVolumeClaimList;
    /** @export {!angular.Resource} - Initialized from binding. */
    this.persistentVolumeClaimListResource;
  }

  /**
   * Returns select id string or undefined if podList or podListResource are not defined.
   * It is needed to enable/disable data select support (pagination, sorting) for particular list.
   *
   * @return {string}
   * @export
   */
  getSelectId() {
    const selectId = 'persistentvolumeclaims';

    if (this.persistentVolumeClaimList !== undefined &&
        this.persistentVolumeClaimListResource !== undefined) {
      return selectId;
    }

    return '';
  }

  /**
   * @return {boolean}
   * @export
   */
  areMultipleNamespacesSelected() {
    return this.kdNamespaceService_.areMultipleNamespacesSelected();
  }
}

/**
 * @return {!angular.Component}
 */
export const persistentVolumeClaimCardListComponent = {
  transclude: {
    // Optional header that is transcluded instead of the default one.
    'header': '?kdHeader',
    // Optional zerostate content that is shown when there are zero items.
    'zerostate': '?kdEmptyListText',
  },
  bindings: {
    'persistentVolumeClaimList': '<',
    'persistentVolumeClaimListResource': '<',
  },
  templateUrl: 'persistentvolumeclaim/list/cardlist.html',
  controller: PersistentVolumeClaimCardListController,
};
