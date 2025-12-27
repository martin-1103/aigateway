import { z } from 'zod'

export const proxySchema = z.object({
  url: z.string().min(1, 'URL is required').url('Must be a valid URL'),
  protocol: z.enum(['http', 'https', 'socks5'], {
    required_error: 'Protocol is required',
  }),
  is_active: z.boolean().default(true),
  max_accounts: z.coerce.number().int().min(1).default(10),
  priority: z.coerce.number().int().min(0).default(0),
  weight: z.coerce.number().int().min(1).default(1),
  max_failures: z.coerce.number().int().min(1).default(5),
})

export type ProxyFormData = z.infer<typeof proxySchema>
