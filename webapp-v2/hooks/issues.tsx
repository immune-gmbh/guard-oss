import useSWR from 'swr';
import { IIncidentsRanking, IRisksRanking, mapRisks, mapIncidents } from './dashboard';

export const RISKS_API_URL = '/v2/risks'
export const INCIDENTS_API_URL = '/v2/incidents'

interface IRisksResponse {
  data: IRisksRanking[]
  isLoading: boolean
  isError: boolean
}

interface IIncidentsResponse {
  data: IIncidentsRanking[]
  isLoading: boolean
  isError: boolean
}

export const useRisks = (): IRisksResponse => {
  const { data, error } = useSWR(RISKS_API_URL);

  return {
    data: data?.risks ? mapRisks(data.risks) : [],
    isError: error,
    isLoading: !error && !data,
  }
}

export const useIncidents = (): IIncidentsResponse => {
  const { data, error } = useSWR(INCIDENTS_API_URL);

  return {
    data: data?.incidents ? mapIncidents(data.incidents) : [],
    isError: error,
    isLoading: !error && !data
  }
}
