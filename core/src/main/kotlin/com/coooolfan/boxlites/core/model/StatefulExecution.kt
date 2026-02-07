package com.coooolfan.boxlites.core.model

data class ExecuteStatefulRequest(
    val name: String?,
    val code: String,
    val leaseSeconds: Long?,
)

data class ExecuteStatefulResult(
    val boxId: String,
    val output: ExecResult,
)
