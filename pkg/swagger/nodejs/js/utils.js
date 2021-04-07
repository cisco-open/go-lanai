export const buildFormData = (data) => {
    const formArr = []

    for (const name in data) {
        if (!data.hasOwnProperty(name)) {
            continue;
        }
        const val = data[name]
        if (val !== undefined && val !== "") {
            formArr.push([name, "=", encodeURIComponent(val).replace(/%20/g,"+")].join(""))
        }
    }
    return formArr.join("&")
}

export function btoa(str) {
    let buffer

    if (str instanceof Buffer) {
        buffer = str
    } else {
        buffer = new Buffer(str.toString(), "utf-8")
    }

    return buffer.toString("base64")
}