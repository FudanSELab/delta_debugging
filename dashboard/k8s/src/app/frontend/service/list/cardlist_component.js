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
export class ServiceCardListController {
  /**
   * @param {!./../../common/namespace/service.NamespaceService} kdNamespaceService
   * @ngInject
   */
  constructor(kdNamespaceService) {
    /** @private {!./../../common/namespace/service.NamespaceService} */
    this.kdNamespaceService_ = kdNamespaceService;
    /** @export {!backendApi.ServiceList} - Initialized from binding. */
    this.serviceList;
    /** @export {!angular.Resource} Initialized from binding. */
    this.serviceListResource;
  }

  /**
   * Returns select id string or undefined if list or list resource are not defined.
   * It is needed to enable/disable data select support (pagination, sorting) for particular list.
   *
   * @return {string}
   * @export
   */
  getSelectId() {
    const selectId = 'services';

    if (this.serviceList !== undefined && this.serviceListResource !== undefined) {
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
 * Definition object for the component that displays service card list.
 *
 * @type {!angular.Component}
 */
export const serviceCardListComponent = {
  transclude: {
    // Optional header that is transcluded instead of the default one.
    'header': '?kdHeader',
    // Optional zerostate content that is shown when there are zero items.
    'zerostate': '?kdEmptyListText',
  },
  templateUrl: 'service/list/cardlist.html',
  controller: ServiceCardListController,
  bindings: {
    /** {!backendApi.ServiceList} */
    'serviceList': '<',
    /** {angular.Resource} */
    'serviceListResource': '<',
    /** {boolean} */
    'selectable': '<',
  },
};
