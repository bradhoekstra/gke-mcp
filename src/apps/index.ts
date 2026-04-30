import { Config } from '../config/config.js';
import { getDropdownTools, getDropdownResources } from './dropdown.js';
import { getChartsTools, getChartsResources } from './charts.js';

export function getAppTools(config: Config) {
  return [
    ...getDropdownTools(config),
    ...getChartsTools(config),
  ];
}

export function getAppResources(config: Config) {
  return [
    ...getDropdownResources(config),
    ...getChartsResources(config),
  ];
}
