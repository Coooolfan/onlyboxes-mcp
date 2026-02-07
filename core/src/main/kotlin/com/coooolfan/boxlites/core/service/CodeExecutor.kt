package com.coooolfan.boxlites.core.service

import com.coooolfan.boxlites.core.model.ExecResult
import com.coooolfan.boxlites.core.model.ExecuteStatefulRequest
import com.coooolfan.boxlites.core.model.ExecuteStatefulResult
import com.coooolfan.boxlites.core.model.RuntimeMetricsView

interface CodeExecutor {
    fun execute(code: String): ExecResult

    fun executeStateful(request: ExecuteStatefulRequest): ExecuteStatefulResult

    fun metrics(): RuntimeMetricsView
}
