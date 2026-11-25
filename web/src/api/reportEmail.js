import service from '@/utils/request'

// The client selects only a server-known report type and optional report ID.
// Recipients, subject and content are deliberately resolved by the backend.
export const sendReportEmail = (data) => service({
  url: '/reportEmail/send',
  method: 'post',
  data
})
