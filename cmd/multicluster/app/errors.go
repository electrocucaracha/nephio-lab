/*
Copyright © 2022

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package app

import "errors"

var (
	// ErrGetName indicates that the cluster name wasn't retrieved.
	ErrGetName = errors.New("unexpected error to get multi-cluster name")

	// ErrGetConfig indicates that the cluster configuration wasn't retrieved.
	ErrGetConfig = errors.New("unexpected error to get multi-cluster configuration")
)
