package com.coooolfan.boxlites.core.port

import com.coooolfan.boxlites.core.model.RuntimeMetricsView

interface BoxFactory {
    fun createStartedBox(): BoxSession

    fun metrics(): RuntimeMetricsView
}
