const demoRows = [
  { name: "IMG_2381.JPG", type: "JPG", before: "3.2 MB", after: "1.4 MB", ratio: "-56%", ok: true },
  { name: "IMG_2382.JPG", type: "JPG", before: "2.9 MB", after: "1.3 MB", ratio: "-55%", ok: true },
  { name: "Poster_A.png", type: "PNG", before: "5.1 MB", after: "2.8 MB", ratio: "-45%", ok: true },
  { name: "Banner_01.webp", type: "WEBP", before: "1.6 MB", after: "1.1 MB", ratio: "-31%", ok: true },
  { name: "Scan_2026_03_01.jpg", type: "JPG", before: "4.3 MB", after: "-", ratio: "-", ok: false },
];

const tbody = document.getElementById("fileRows");

demoRows.forEach((item) => {
  const tr = document.createElement("tr");
  tr.innerHTML = `
    <td>${item.name}</td>
    <td>${item.type}</td>
    <td>${item.before}</td>
    <td>${item.after}</td>
    <td>${item.ratio}</td>
    <td class="${item.ok ? "status-ok" : "status-fail"}">${item.ok ? "成功" : "失败（重试3次后）"}</td>
  `;
  tbody.appendChild(tr);
});

const simulation = {
  totalFiles: 100,
  doneFiles: 72,
  concurrency: 4,
  currentElapsedSec: 3,
  currentTotalSec: 7,
  perFileSeconds: 7,
  avgPerFileSeconds: 0.56,
};

const overallProgressText = document.getElementById("overallProgressText");
const overallProgressCount = document.getElementById("overallProgressCount");
const overallProgressBar = document.getElementById("overallProgressBar");
const currentProgressText = document.getElementById("currentProgressText");
const currentProgressBar = document.getElementById("currentProgressBar");
const elapsedText = document.getElementById("elapsedText");
const remainingText = document.getElementById("remainingText");
const kpiSpeed = document.getElementById("kpiSpeed");
const kpiEta = document.getElementById("kpiEta");
const kpiConcurrency = document.getElementById("kpiConcurrency");

function formatClock(seconds) {
  const safe = Math.max(0, Math.floor(seconds));
  const hh = String(Math.floor(safe / 3600)).padStart(2, "0");
  const mm = String(Math.floor((safe % 3600) / 60)).padStart(2, "0");
  const ss = String(safe % 60).padStart(2, "0");
  return `${hh}:${mm}:${ss}`;
}

function computeEtaSeconds() {
  const remainingFiles = simulation.totalFiles - simulation.doneFiles;
  const currentRemaining = Math.max(0, simulation.currentTotalSec - simulation.currentElapsedSec);
  const queuedFiles = Math.max(0, remainingFiles - 1);
  return currentRemaining + Math.ceil((queuedFiles * simulation.avgPerFileSeconds) / simulation.concurrency);
}

function renderSimulation() {
  const overallPercent = (simulation.doneFiles / simulation.totalFiles) * 100;
  const currentPercent = (simulation.currentElapsedSec / simulation.currentTotalSec) * 100;
  const remainSec = Math.max(0, simulation.currentTotalSec - simulation.currentElapsedSec);
  const etaSec = computeEtaSeconds();

  overallProgressText.textContent = `总进度 ${overallPercent.toFixed(0)}%`;
  overallProgressCount.textContent = `${simulation.doneFiles} / ${simulation.totalFiles}`;
  overallProgressBar.style.width = `${overallPercent}%`;

  currentProgressText.textContent = `当前文件进度 ${currentPercent.toFixed(0)}%`;
  currentProgressBar.style.width = `${currentPercent}%`;
  elapsedText.textContent = `已运行 ${formatClock(simulation.currentElapsedSec)}`;
  remainingText.textContent = `预计剩余：${formatClock(remainSec)}`;

  kpiConcurrency.textContent = String(simulation.concurrency);
  kpiSpeed.textContent = `${(1 / simulation.avgPerFileSeconds).toFixed(1)} 张/秒`;
  kpiEta.textContent = formatClock(etaSec);
}

function tick() {
  simulation.currentElapsedSec += 1;

  if (simulation.currentElapsedSec >= simulation.currentTotalSec) {
    simulation.doneFiles = Math.min(simulation.totalFiles, simulation.doneFiles + 1);
    simulation.currentElapsedSec = 0;
    simulation.currentTotalSec = simulation.perFileSeconds + Math.floor(Math.random() * 4);
  }

  if (simulation.doneFiles >= simulation.totalFiles) {
    simulation.doneFiles = simulation.totalFiles;
    simulation.currentElapsedSec = simulation.currentTotalSec;
    renderSimulation();
    return;
  }

  renderSimulation();
  window.setTimeout(tick, 1000);
}

renderSimulation();
window.setTimeout(tick, 1000);
