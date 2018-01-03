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

import {EditResourceController} from 'common/resource/editresource_controller';
import resourceModule from 'common/resource/module';

describe('Edit resource controller', () => {
  /** @type {!common/resource/editresource_controller.EditResourceController} */
  let ctrl;
  /** @type {!md.$dialog} */
  let mdDialog;
  /** @type {!angular.$httpBackend} */
  let httpBackend;
  let testResourceUrl = 'api/v1/testurl';

  beforeEach(() => {
    angular.mock.module(resourceModule.name, ($provide) => {

      let localizerService = {localize: function() {}};

      $provide.value('localizerService', localizerService);
    });

    angular.mock.inject(($controller, $mdDialog, $httpBackend) => {
      ctrl = $controller(EditResourceController, {
        resourceKindName: 'My Resource',
        resourceUrl: testResourceUrl,
      });
      mdDialog = $mdDialog;
      httpBackend = $httpBackend;
    });
  });

  it('should edit resource', () => {
    spyOn(mdDialog, 'hide');
    ctrl.update();
    let data = {'foo': 'bar'};
    httpBackend.expectGET(testResourceUrl).respond(200, data);
    httpBackend.expectPUT(testResourceUrl).respond(200, {ok: 'ok'});
    httpBackend.flush();
    expect(mdDialog.hide).toHaveBeenCalled();
  });

  it('should propagate errors', () => {
    spyOn(mdDialog, 'cancel');
    ctrl.update();
    let data = {'foo': 'bar'};
    httpBackend.expectGET(testResourceUrl).respond(200, data);
    httpBackend.expectPUT(testResourceUrl).respond(500, {err: 'err'});
    httpBackend.flush();
    expect(mdDialog.cancel).toHaveBeenCalled();
  });

  it('should cancel', () => {
    spyOn(mdDialog, 'cancel');
    ctrl.cancel();
    expect(mdDialog.cancel).toHaveBeenCalled();
  });
});
