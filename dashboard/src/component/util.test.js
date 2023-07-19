import { getDefaultFilters, inShort } from "./util";


test("default filters", () => {
    const list = [{
        "id": 1,
    }, {
        "id": 2,
    }]

    const filter = getDefaultFilters(list, (e) => e.id);
    expect(filter).toEqual([{
        "text": 1,
        "value": 1,
    }, {
        "text": 2,
        "value": 2,
    }])

    const filter2 = getDefaultFilters(list, (e) => e.id, (e) => e + 1);
    expect(filter2).toEqual([{
        "text": 2,
        "value": 1,
    }, {
        "text": 3,
        "value": 2,
    }])
})


test("in short", () => {
    const case1 = "1234567890"
    expect(inShort(case1)).toEqual(case1)
    const case2 = "12345678901234567890"
    expect(inShort(case2, 3, 4)).toEqual("123...7890")
    expect(inShort(case2, 3, 4, '..')).toEqual("123..7890")

    const case3 = 'bafyreiali7lqcrskxz7c2wgo5uevbdb7l4g7fu77a43n7cxyla7cesbcha'
    expect(inShort(case3, 10, 10)).toEqual('bafyreiali...la7cesbcha')
})
