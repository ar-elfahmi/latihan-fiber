const token = "TOKEN_DISINI"

// Simulasi 10 request serentak
const results = []
for (let i = 0; i < 10; i++) {
    const result = db.student_portfolios.findOneAndUpdate(
        {
            student_id: 1,
            "projects.19": { $exists: false }
        },
        {
            $push: {
                projects: {
                    _id: new ObjectId(),
                    title: `Concurrent Project ${i+1}`,
                    technologies: ["Go"],
                    url: `https://github.com/example/concurrent-${i+1}`,
                    year: 2026
                }
            }
        },
        { returnDocument: "after" }
    )
    results.push(result ? "201 OK" : "409 PROJECT_LIMIT")
}

print("Hasil:")
results.forEach((r, i) => print(`Request ${i+1}: ${r}`))
print(`\nTotal berhasil: ${results.filter(r => r.includes("OK")).length}`)
print(`Total ditolak: ${results.filter(r => r.includes("409")).length}`)
print(`Total projects akhir: ${db.student_portfolios.findOne({student_id: 1}).projects.length}`)
