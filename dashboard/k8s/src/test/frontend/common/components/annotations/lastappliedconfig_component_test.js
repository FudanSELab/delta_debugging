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

import module from 'common/components/annotations/module';

describe('last applied config component', () => {
  /** @type {!common/components/annotations/lastappliedconfiguration_component.Controller} */
  let ctrl;
  /** @type {!md.$mdDialog} */
  let mdDialog;

  beforeEach(() => {
    angular.mock.module(module.name);

    angular.mock.inject(($componentController, $mdDialog) => {
      ctrl = $componentController('kdLastAppliedConfiguration');
      mdDialog = $mdDialog;
    });
  });

  it('should open the dialog with details', () => {
    spyOn(mdDialog, 'show');
    ctrl.openDetails();
    expect(mdDialog.show).toHaveBeenCalled();
  });
});
