import { Fil, Power } from './util';
import { separateValueAndUnit } from './util';

test('test Fil ', () => {
    let res = Fil(1);
    expect(res).toBe('1.00 aFIL');
    res = Fil(1000);
    expect(res).toBe('1.00 fFIL');
    res = Fil(1000000);
    expect(res).toBe('1.00 pFIL');
    res = Fil(1000000000000000000);
    expect(res).toBe('1.00 FIL');

    res = Fil(-1000000000000000000);
    expect(res).toBe('-1.00 FIL');
});


test('test Power ', () => {
    let res = Power(1);
    expect(res).toBe('1.00 iB');
    res = Power(1024);
    expect(res).toBe('1.00 KiB');
    res = Power(1024 * 1024);
    expect(res).toBe('1.00 MiB');
    res = Power(1024 * 1024 * 1024 * 1024 * 1024 * 1024);
    expect(res).toBe('1.00 EiB');
});


test("date transfer", () => {
    let a = new Date("2023-02-23 06:53:20")
    let b = new Date("2023-02-22 06:53:20")

    expect(a > b).toBe(true)
})


test('separateValueAndUnit', () => {
    const correctCases = [
        { input: '10px', expected: [10, 'px'] },
        { input: '10.5px', expected: [10.5, 'px'] },
        { input: '10.5 px', expected: [10.5, 'px'] },
        { input: '10.5', expected: [10.5, ''] },
    ]
    correctCases.forEach((c) => {
        expect(separateValueAndUnit(c.input)).toEqual(c.expected)
    })

    const errorCases = [
        { input: 'px' },
        { input: '10.5px10.5px' },
        { input: '10.5.5px' },
        { input: '10.5.5' },
        { input: "asd" },
    ]
    errorCases.forEach((c) => {
        try {
            expect(() => {
                const res = separateValueAndUnit(c.input)
                console.log(res, c.input);
            }).toThrow("Invalid string format");
        } catch (error) {
            throw new Error(`${c.input}: ${error}`);
        }
    })
})
