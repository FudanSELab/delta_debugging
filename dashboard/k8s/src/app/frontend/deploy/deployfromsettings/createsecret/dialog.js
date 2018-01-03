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
export class CreateSecretController {
  /**
   * @param {!md.$dialog} $mdDialog
   * @param {!angular.$log} $log
   * @param {!angular.$resource} $resource
   * @param {!../../../common/errorhandling/dialog.ErrorDialog} errorDialog
   * @param {string} namespace
   * @param {!../../../common/csrftoken/service.CsrfTokenService} kdCsrfTokenService
   * @param {string} kdCsrfTokenHeader
   * @ngInject
   */
  constructor(
      $mdDialog, $log, $resource, errorDialog, namespace, kdCsrfTokenService, kdCsrfTokenHeader) {
    /** @private {!md.$dialog} */
    this.mdDialog_ = $mdDialog;

    /** @private {!angular.$log} */
    this.log_ = $log;

    /** @private {!angular.$resource} */
    this.resource_ = $resource;

    /** @private {!../../../common/errorhandling/dialog.ErrorDialog} */
    this.errorDialog_ = errorDialog;

    /** @export {!angular.FormController} */
    this.secretForm;

    /**
     * The current selected namespace, initialized from the scope.
     * @export {string}
     */
    this.namespace = namespace;

    /** @export {string} */
    this.secretName;

    /**
     * The Base64 encoded data for the ImagePullSecret.
     * @export {string}
     */
    this.data;

    /**
     * Max-length validation rule for secretName.
     * @export {number}
     */
    this.secretNameMaxLength = 253;

    /**
     * Pattern validation rule for secretName.
     * @export {!RegExp}
     */
    this.secretNamePattern =
        new RegExp('^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$');

    /**
     * Pattern validating if the secret data is Base64 encoded.
     * @export {!RegExp}
     */
    this.dataPattern =
        new RegExp('^([0-9a-zA-Z+/]{4})*(([0-9a-zA-Z+/]{2}==)|([0-9a-zA-Z+/]{3}=))?$');

    /** @private {!angular.$q.Promise} */
    this.tokenPromise = kdCsrfTokenService.getTokenForAction('secret');

    /** @private {string} */
    this.csrfHeaderName_ = kdCsrfTokenHeader;
  }

  /**
   * Cancels the create secret form.
   * @export
   */
  cancel() {
    this.mdDialog_.cancel();
  }

  /**
   * Creates new secret based on the state of the controller.
   * @export
   */
  createSecret() {
    if (!this.secretForm.$valid) return;

    /** @type {!backendApi.SecretSpec} */
    let secretSpec = {
      name: this.secretName,
      namespace: this.namespace,
      data: this.data,
    };
    this.tokenPromise.then(
        (token) => {
          /** @type {!angular.Resource} */
          let resource = this.resource_(
              `api/v1/secret/`, {},
              {save: {method: 'POST', headers: {[this.csrfHeaderName_]: token}}});

          resource.save(
              secretSpec,
              (savedConfig) => {
                this.log_.info('Successfully created secret:', savedConfig);
                this.mdDialog_.hide(this.secretName);
              },
              (err) => {
                this.mdDialog_.hide();
                this.errorDialog_.open('Error creating secret', err.data);
                this.log_.info('Error creating secret:', err);
              });
        },
        (err) => {
          this.mdDialog_.hide();
          this.errorDialog_.open('Error creating secret', err.data);
          this.log_.info('Error creating secret:', err);
        });
  }
}

/**
 * @param {!md.$dialog} mdDialog
 * @param {!angular.Scope.Event} event
 * @param {string} namespace
 * @return {!angular.$q.Promise}
 */
export default function showCreateSecretDialog(mdDialog, event, namespace) {
  return mdDialog.show({
    controller: CreateSecretController,
    controllerAs: '$ctrl',
    clickOutsideToClose: true,
    targetEvent: event,
    templateUrl: 'deploy/deployfromsettings/createsecret/createsecret.html',
    locals: {
      'namespace': namespace,
    },
  });
}
