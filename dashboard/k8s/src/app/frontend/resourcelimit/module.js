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

import chromeModule from '../chrome/module';
import componentsModule from '../common/components/module';
import filtersModule from '../common/filters/module';
import eventsModule from '../events/module';
import {resourceLimitsComponent} from './detail/detail_component';

/**
 * Angular module for the Limit Range details view.
 *
 * The view shows detailed view of a Limit Range.
 */
export default angular
    .module(
        'kubernetesDashboard.resourceLimit',
        [
          'ngMaterial',
          'ngResource',
          'ui.router',
          componentsModule.name,
          filtersModule.name,
          eventsModule.name,
          chromeModule.name,
        ])
    .component('kdResourceLimits', resourceLimitsComponent);
