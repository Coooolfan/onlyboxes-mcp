package com.coooolfan.onlyboxes.core.model

data class ExecuteStatefulRequest(
    val ownerToken: String,
    val name: String?,
    val code: String,
    val leaseSeconds: Long,
)

data class ExecuteStatefulResult(
    val boxId: String?,
    val destroyed: Boolean,
    val remainingDestroySeconds: Long,
    val output: ExecResult,
)
