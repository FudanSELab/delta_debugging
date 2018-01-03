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

import {infoCardComponent} from './infocard_component';
import {infoCardEntryComponent} from './infocardentry_component';
import {infoCardSectionComponent} from './infocardsection_component';

/**
 * Module containing common components for cards that can carry any content.
 */
export default angular
    .module(
        'kubernetesDashboard.common.components.infocard',
        [
          'ngMaterial',
        ])
    .component('kdInfoCard', infoCardComponent)
    .component('kdInfoCardEntry', infoCardEntryComponent)
    .component('kdInfoCardSection', infoCardSectionComponent);
