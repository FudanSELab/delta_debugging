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

import settingsServiceModule from 'common/settings/module';
import module from 'discovery/module';
import {resolveResource} from 'discovery/stateconfig';

describe('StateConfig for discovery list', () => {
  /** @type {!common/dataselect/dataselect_service.DataSelectService} */
  let kdDataSelectService;

  beforeEach(() => {
    angular.mock.module(module.name);
    angular.mock.module(settingsServiceModule.name);
    angular.mock.inject((_kdDataSelectService_) => {
      kdDataSelectService = _kdDataSelectService_;
    });
  });

  it('should resolve discovery', angular.mock.inject(($q) => {
    let promise = $q.defer().promise;

    let resource = jasmine.createSpyObj('$resource', ['get']);
    resource.get.and.callFake(function() {
      return {$promise: promise};
    });

    let actual = resolveResource(resource, {namespace: 'foo-ns'}, kdDataSelectService);

    expect(resource.get)
        .toHaveBeenCalledWith(kdDataSelectService.getDefaultResourceQuery('foo-ns'));
    expect(actual).toBe(promise);
  }));
});
