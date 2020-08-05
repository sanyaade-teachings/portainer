import angular from 'angular';
import _ from 'lodash-es';
import PortainerError from 'Portainer/error';
import KubernetesStorageClassConverter from 'Kubernetes/converters/storageClass';

class KubernetesStorageService {
  /* @ngInject */
  constructor($async, KubernetesStorage) {
    this.$async = $async;
    this.KubernetesStorage = KubernetesStorage;

    this.getAsync = this.getAsync.bind(this);
    this.patchAsync = this.patchAsync.bind(this);
  }

  /**
   * GET
   */
  async getAsync(endpointId) {
    try {
      const params = {
        endpointId: endpointId,
      };
      const classes = await this.KubernetesStorage().get(params).$promise;
      const res = _.map(classes.items, (item) => KubernetesStorageClassConverter.apiToStorageClass(item));
      return res;
    } catch (err) {
      throw new PortainerError('Unable to retrieve storage classes', err);
    }
  }

  get(endpointId) {
    return this.$async(this.getAsync, endpointId);
  }

  /**
   * PATCH
   */
  async patchAsync(oldStorageClass, newStorageClass) {
    try {
      const params = {
        id: newStorageClass.Name,
      };
      const payload = KubernetesStorageClassConverter.patchPayload(oldStorageClass, newStorageClass);
      await this.KubernetesStorage().patch(params, payload).$promise;
    } catch (err) {
      throw new PortainerError('Unable to patch storage class', err);
    }
  }

  patch(oldStorageClass, newStorageClass) {
    return this.$async(this.patchAsync, oldStorageClass, newStorageClass);
  }
}

export default KubernetesStorageService;
angular.module('portainer.kubernetes').service('KubernetesStorageService', KubernetesStorageService);
