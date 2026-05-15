// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

package packager

import (
	"context"

	"github.com/zarf-dev/zarf/src/api/v1alpha1"
	"github.com/zarf-dev/zarf/src/pkg/packager/load"
)

type CheckChartsOptions struct {
	Flavor    string
	CachePath string
}

type ChartScan struct {
	ComponentName string
	Charts        []v1alpha1.ZarfChart
}

func FindLatestCharts(ctx context.Context, packagePath string, opts CheckChartsOptions) ([]ChartScan, error) {
	definitionOptions := load.DefinitionOptions{
		Flavor:           opts.Flavor,
		CachePath:        opts.CachePath,
		SkipVersionCheck: true,
	}

	pkg, err := load.PackageDefinition(ctx, packagePath, definitionOptions)

	if err != nil {
		return nil, err
	}

	var chartScans []ChartScan

	for _, component := range pkg.Components {
		scan := ChartScan{
			ComponentName: component.Name,
		}

		updatedHelmRepositoryCharts, err := updateHelmRepositoryChartsToLatest(scan.Charts)
		if err != nil {
			return nil, err
		}

		scan.Charts = append(scan.Charts, updatedHelmRepositoryCharts...)

		chartScans = append(chartScans, scan)
	}

	return chartScans, nil
}

func updateHelmRepositoryChartsToLatest(charts []v1alpha1.ZarfChart) ([]v1alpha1.ZarfChart, error) {
	var helmRepositoryCharts []v1alpha1.ZarfChart

	for _, chart := range charts {
		helmRepositoryCharts = append(helmRepositoryCharts, chart)

	}

	return helmRepositoryCharts, nil
}
