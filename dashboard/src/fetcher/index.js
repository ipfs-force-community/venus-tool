import { ApiBase } from "../global"
import axios from "axios";
import QueryString from "qs";
import qs from "qs";
import useSWR from "swr";


export const rel = relative => `${ApiBase}${relative}`
export const DefaultFetcher = function (url) {
    return axios.get(url).then(res => {
        return res.data
    }).catch(rawErr => {
        const err = rawErr.response
        throw err
    })
}
const fetcherGetWithParams = function ([url, params]) {
    if (params instanceof Error) {
        console.warn(params.message);
        return Promise.reject(params)
    }
    return axios.get(url, { params: params, paramsSerializer: p => qs.stringify(p, { arrayFormat: "repeat" }) }).then(res => {
        return res.data
    }).catch(rawErr => {
        const err = rawErr.response
        throw err
    })
}

const fetchOrError = function (maybeError) {
    if (maybeError instanceof Error) {
        console.warn(maybeError.message);
        return () => Promise.reject(maybeError)
    } else {
        return DefaultFetcher
    }
}


export const useWallets = function () {
    return useSWR(rel("/wallet/list"), { fallbackData: [] })
}

export const useMsgs = function (wallet) {
    const params = wallet ? {
        "From": `"${wallet}"`,
        "Offset": 0,
        "Limit": 200,
    } : new Error("useMsgs: wallet is null")
    return useSWR([rel("/msg/query"), params], fetcherGetWithParams, { fallbackData: [] })
}

export const useMsgInfo = function (id) {
    const params = id ? {
        "ID": id,
    } : new Error("useMsgInfo: id is null")
    return useSWR([rel(`/msg/`), params], fetcherGetWithParams)
}

export const useThreads = function () {
    return useSWR(rel("/thread/list"), { fallbackData: [] })
}

export const useMiners = function () {
    return useSWR(rel("/miner/list"), { fallbackData: [] })
}

export const useWalletInfo = function (wallet) {
    return useSWR(rel(`/addr/info/"${wallet}"`), fetchOrError(wallet ? null : new Error("useWalletInfo: wallet is null")))
}

export const useMinerInfo = function (miner) {
    return useSWR(rel(`/miner/info/"${miner}"`), fetchOrError(miner ? null : new Error("useMinerInfo: miner is null")))
}

export const useMinerInfos = function () {
    let ret = []

    const fetcher = () => axios.get(rel("/miner/list")).then(res => {
        let miners = res.data
        let promises = miners.map(miner => {
            return axios.get(rel(`/miner/info/"${miner}"`)).then(res => {
                return res.data
            })
        })
        return axios.all(promises).then(res => {
            res.forEach(data => {
                ret = ret.concat(data)
            })
            return ret
        })
    })

    return useSWR(["miner_infos"], fetcher, { fallbackData: [] })
}

export const useWalletInfos = function () {
    let ret = []

    const fetcher = () => axios.get(rel("/wallet/list")).then(res => {
        let wallets = res.data
        let promises = wallets.map(wallet => {
            return axios.get(rel(`/addr/info/"${wallet}"`)).then(res => {
                return res.data
            })
        })
        return axios.all(promises).then(res => {
            res.forEach(data => {
                ret = ret.concat(data)
            })
            return ret
        })
    })

    return useSWR(["wallet_infos"], fetcher, { fallbackData: [] })
}


export const useSectorSum = function ({ miner }) {
    const params = miner ? {
        "Address": `"${miner}"`,
    } : new Error("useSectorSum: miner is null")
    return useSWR([rel("/sector/sum"), params], fetcherGetWithParams)
}

export const useSectors = function ({ miner, pageIndex, pageSize }) {
    const params = miner ? {
        "Miner": `"${miner}"`,
        "PageIndex": pageIndex,
        "PageSize": pageSize,
    } : new Error("useSectors: miner is null")
    return useSWR([rel("/sector/list"), params], fetcherGetWithParams)
}

export const useDeals = function ({ miner }) {
    return useSWR(rel(`/deal/storage/"${miner}"`), fetchOrError(miner ? null : new Error("useDeals: miner is null")), { fallbackData: [] })
}

export const useDealInfo = function (pCid) {
    return useSWR(rel(`/deal/storage/info/${pCid}`))
}

export const useAllDeals = function (miners) {
    let ret = []
    let promises = miners.map(miner => {
        return axios.get(rel(`/deal/storage/"${miner}"`)).then(res => {
            return res.data
        })
    })
    const fetcher = () => axios.all(promises).then(res => {
        res.forEach(data => {
            ret = ret.concat(data)
        })
        return ret
    })

    return useSWR(["all_deals", miners], fetcher, { fallbackData: [] })
}


export const useMsgsByUpdate = function ({ updateBefore }) {
    const params = {
        "ByUpdateAt": updateBefore.toISOString(),
        "Limit": 500,
    }

    return useSWR([rel("/msg/query"), params], fetcherGetWithParams)
}

export const useBlockList = function (miners) {

    const params = miners ? {
        "Miner": miners.map(miner => `"${miner}"`),
        "Limit": 200,
    } : new Error("useBlockList: miners is null")
    return useSWR([rel("/minedblock/list"), params], fetcherGetWithParams)
}



export const useSearch = function (keyword) {
    const params = keyword ? {
        "Key": keyword,
    } : new Error("useSearch: keyword is null")
    return useSWR([rel("/search"), params], fetcherGetWithParams)
}


export const Search = function (keyword) {
    const params = {
        "Key": keyword,
    }
    return fetcherGetWithParams([rel("/search"), params])
}


export const SendMsg = function (data) {
    return axios.post(rel("/msg/send"), data).then(res => {
        return res.data
    }).catch(rawErr => {
        const err = rawErr.response
        throw err
    })
}


export const MarkBad = function (data) {
    return axios.post(rel("/msg/markbad"), data).then(res => {
        return res.data
    }).catch(rawErr => {
        const err = rawErr.response
        throw err
    })
}
