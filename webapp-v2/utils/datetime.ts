import { Translate } from "next-translate"

export const timeAgoFormat = (date: string, t: Translate): string => {
  const dateFormat = new Date(date)
  const now = new Date()

  const diffMiliseconds = +now - +dateFormat
  const minutesAgo = diffMiliseconds / (1000 * 60)

  if (minutesAgo < 60) return t('utils:date.minutesAgo', { count: Math.floor(minutesAgo) })

  const hoursAgo = minutesAgo / 60
  if (hoursAgo < 24) return t('utils:date.hoursAgo', { count: Math.floor(hoursAgo) })

  const daysAgo = hoursAgo / 24
  if (daysAgo < 30) return t('utils:date.daysAgo', { count: Math.floor(daysAgo) })

  const weeksAgo = daysAgo / 7
  if (weeksAgo < 4.3) return t('utils:date.weeksAgo', { count: Math.floor(weeksAgo) })

  const monthsAgo = weeksAgo / 4.3
  if (monthsAgo < 12) return t('utils:date.monthsAgo', { count: Math.floor(monthsAgo) })

  return t('utils:date.moreThanYearAgo')
}


export const rfcToStringFormat = (date: string): string => {
  const dateFormat = new Date(date)
  return dateFormat.toUTCString()
}
