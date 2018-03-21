"""Airfow DAG used is the daily release pipeline.

Copyright 2017 Istio Authors. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
"""

from airflow import DAG
import istio_common_dag

dag, copy_files = istio_common_dag.MakeCommonDag(
    name='istio_daily_release', schedule_interval='15 9 * * *')

mark_complete = istio_common_dag.MakeMarkComplete(dag)

copy_files >> mark_complete

dag
