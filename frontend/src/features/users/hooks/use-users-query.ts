import { useQuery } from '@tanstack/react-query'
import { getUsers } from '../api'

export const usersQueryKey = ['users']

export function useUsersQuery() {
  return useQuery({
    queryKey: usersQueryKey,
    queryFn: getUsers,
  })
}
