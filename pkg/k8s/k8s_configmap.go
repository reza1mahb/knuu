package k8s

import (
	"context"

	v1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Client) GetConfigMap(ctx context.Context, name string) (*v1.ConfigMap, error) {
	cm, err := c.clientset.CoreV1().ConfigMaps(c.namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, ErrGettingConfigmap.WithParams(name).Wrap(err)
	}

	return cm, nil
}

func (c *Client) ConfigMapExists(ctx context.Context, name string) (bool, error) {
	_, err := c.clientset.CoreV1().ConfigMaps(c.namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrs.IsNotFound(err) {
			return false, nil
		}
		return false, ErrGettingConfigmap.WithParams(name).Wrap(err)
	}

	return true, nil
}

func (c *Client) CreateConfigMap(
	ctx context.Context,
	name string,
	labels, data map[string]string,
) (*v1.ConfigMap, error) {
	exists, err := c.ConfigMapExists(ctx, name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrConfigmapAlreadyExists.WithParams(name)
	}

	cm, err := prepareConfigMap(c.namespace, name, labels, data)
	if err != nil {
		return nil, err
	}

	created, err := c.clientset.CoreV1().ConfigMaps(c.namespace).Create(ctx, cm, metav1.CreateOptions{})
	if err != nil {
		return nil, ErrCreatingConfigmap.WithParams(name).Wrap(err)
	}

	return created, nil
}

func (c *Client) DeleteConfigMap(ctx context.Context, name string) error {
	exists, err := c.ConfigMapExists(ctx, name)
	if err != nil {
		return err
	}
	if !exists {
		return ErrConfigmapDoesNotExist.WithParams(name)
	}

	err = c.clientset.CoreV1().ConfigMaps(c.namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return ErrDeletingConfigmap.WithParams(name).Wrap(err)
	}

	return nil
}

func prepareConfigMap(
	namespace, name string,
	labels, data map[string]string,
) (*v1.ConfigMap, error) {
	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Data: data,
	}
	return cm, nil
}
