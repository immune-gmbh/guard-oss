import { useState } from 'react';
import useSWR from 'swr';
import { convertToCamelRecursive } from 'utils/case';

export const DASHBOARD_API_URL = '/v2/dashboard';

export interface IDashboardData {
  deviceStats: IDeviceStats;
  incidents: { count: number; devices: number };
  risks: IRisksRanking[];
}

interface IDashboardResponse {
  data: IDashboardData;
  isLoading: boolean;
  isError: boolean;
  loadData: () => void;
}

export interface IRisksRanking {
  issueId: string;
  count: number;
}

export interface IIncidentsRanking {
  issueId: string;
  count: number;
  timestamp: string;
}

export type IIssueRanking = IRisksRanking | IIncidentsRanking;

export interface IDeviceStats {
  numTrusted: number;
  numWithIncident: number;
  numAtRisk: number;
  numUnresponsive: number;
}

export const mapRisks = (
  risksList: Array<{ issue_type_id: string; num_occurences: number }>,
): IRisksRanking[] =>
  risksList.map((risk) => ({
    issueId: risk.issue_type_id,
    count: risk.num_occurences,
  }));

export const mapIncidents = (
  incidentsList: Array<{
    issue_type_id: string;
    devices_affected: number;
    latest_occurence: string;
  }>,
): IIncidentsRanking[] =>
  incidentsList.map((incident) => ({
    issueId: incident.issue_type_id,
    count: incident.devices_affected,
    timestamp: incident.latest_occurence,
  }));

export const useDashboard = (loadOnInit = false): IDashboardResponse => {
  const [ load, setLoad ] = useState(loadOnInit)
  const { data, error, mutate } = useSWR(() => load ? DASHBOARD_API_URL : null);

  return {
    data: {
      deviceStats: convertToCamelRecursive(data?.device_stats) as IDeviceStats,
      incidents: {
        count: data?.incident_count,
        devices: data?.incident_dev_count,
      },
      risks: data?.risks ? mapRisks(data.risks) : [],
    },
    isLoading: !error && !data,
    isError: error,
    loadData: async () => {
      setLoad(true)
      await mutate(DASHBOARD_API_URL)
    }
  };
};
