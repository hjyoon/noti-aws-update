/* ========= 상수·상태 ========= */
const PAGE_SIZE = 12,
  SIDEBAR_MIN = 240,
  SIDEBAR_MAX = 600;
let lastSidebarWidth = 320,
  isSidebarCollapsed = false;
let currentPage = 1,
  selectedTags = [],
  tagSearchKeyword = "",
  tagPage = 1,
  tagTotalPage = 1,
  tagPageSize = 30,
  tagListData = [],
  newsSearchKeyword = "",
  cardItems = [],
  totalCount = 0,
  isLoading = false,
  isEndOfList = false;

/* ========= 사이드바 모바일 ========= */
function openSidebar() {
  const sb = document.getElementById("sidebar");
  sb.classList.remove("closed", "hidden");
  sb.style.width = sb.style.padding = sb.style.overflow = "";
  sb.classList.add(
    "fixed",
    "left-0",
    "top-[65px]",
    "h-[calc(100vh-65px)]",
    "bg-white",
    "w-[80vw]",
    "max-w-[400px]",
    "z-30",
  );
  document.getElementById("sidebarOverlay").classList.remove("hidden");
}
function closeSidebar() {
  const sb = document.getElementById("sidebar");
  sb.classList.add("hidden");
  sb.classList.remove(
    "fixed",
    "left-0",
    "top-[65px]",
    "h-[calc(100vh-65px)]",
    "bg-white",
    "w-[80vw]",
    "max-w-[400px]",
    "z-30",
  );
  document.getElementById("sidebarOverlay").classList.add("hidden");
}

/* ========= 데스크톱 토글 ========= */
function toggleSidebarDesktop() {
  const sb = document.getElementById("sidebar"),
    handle = document.getElementById("sidebarResizeHandle");
  if (isSidebarCollapsed) {
    sb.classList.remove("closed");
    sb.style.borderRight = "1px solid rgb(229 231 235)";
    handle.style.display = "block";
    isSidebarCollapsed = false;
  } else {
    lastSidebarWidth = sb.offsetWidth;
    sb.classList.add("closed");
    sb.style.borderRight = "none";
    handle.style.display = "none";
    isSidebarCollapsed = true;
  }
}

/* ========= 리사이즈 핸들 ========= */
function initSidebarResize() {
  const sb = document.getElementById("sidebar"),
    h = document.getElementById("sidebarResizeHandle");
  if (!h) return;
  let sx = 0,
    sw = 0,
    res = false;
  h.addEventListener("mousedown", (e) => {
    if (isSidebarCollapsed) return;
    res = true;
    sx = e.clientX;
    sw = sb.offsetWidth;
    document.body.style.userSelect = "none";
  });
  document.addEventListener("mousemove", (e) => {
    if (!res) return;
    let w = sw + (e.clientX - sx);
    w = Math.max(SIDEBAR_MIN, Math.min(SIDEBAR_MAX, w));
    sb.style.width = w + "px";
  });
  document.addEventListener("mouseup", () => {
    if (res) {
      res = false;
      lastSidebarWidth = sb.offsetWidth;
      document.body.style.userSelect = "";
    }
  });
}

/* ========= DOMContentLoaded ========= */
document.addEventListener("DOMContentLoaded", () => {
  loadTagList();
  resetAndLoadNews();
  initSidebarResize();

  document.getElementById("sidebarOpenBtn").onclick = () => {
    window.innerWidth < 768 ? openSidebar() : toggleSidebarDesktop();
  };
  document.getElementById("sidebarCollapseBtnDesktop").onclick =
    toggleSidebarDesktop;
  document.getElementById("sidebarCloseBtn").onclick = closeSidebar;
  document.getElementById("sidebarOverlay").onclick = closeSidebar;

  document.getElementById("searchBtn").onclick = () => {
    newsSearchKeyword = document.getElementById("searchInput").value.trim();
    resetAndLoadNews();
  };
  document.getElementById("tagSearchBtn").onclick = () => {
    tagSearchKeyword = document.getElementById("tagSearch").value.trim();
    tagPage = 1;
    loadTagList();
  };
  document.getElementById("tagPrevBtn").onclick = () => {
    if (tagPage > 1) {
      tagPage--;
      loadTagList();
    }
  };
  document.getElementById("tagNextBtn").onclick = () => {
    if (tagPage < tagTotalPage) {
      tagPage++;
      loadTagList();
    }
  };
  document.getElementById("searchInput").addEventListener("keyup", (e) => {
    if (e.key === "Enter") {
      newsSearchKeyword = e.target.value.trim();
      resetAndLoadNews();
    }
  });
  document.getElementById("tagSearch").addEventListener("keyup", (e) => {
    if (e.key === "Enter") {
      tagSearchKeyword = e.target.value.trim();
      tagPage = 1;
      loadTagList();
    }
  });

  document
    .getElementById("mainScrollArea")
    .addEventListener("scroll", onMainScroll);

  /* 브라우저 크기 변화 대응 */
  window.addEventListener("resize", () => {
    const sb = document.getElementById("sidebar");
    const handle = document.getElementById("sidebarResizeHandle");

    if (window.innerWidth >= 768) {
      // 모바일 클래스 정리
      closeSidebar();

      // 접힘 상태 유지
      if (isSidebarCollapsed) {
        sb.classList.add("closed");
        sb.style.borderRight = "none";
        handle.style.display = "none";
      } else {
        sb.classList.remove("closed");
        sb.style.borderRight = "1px solid rgb(229 231 235)";
        handle.style.display = "block";
      }
    }
  });
});

/* ========= 태그 API ========= */
function loadTagList() {
  let url = `/api/tags?limit=${tagPageSize}&offset=${(tagPage - 1) * tagPageSize}`;
  if (tagSearchKeyword) url += `&name=${encodeURIComponent(tagSearchKeyword)}`;
  fetch(url)
    .then((r) => r.json())
    .then((data) => {
      tagListData = data.items;
      tagTotalPage = data.total_page || 1;
      renderTagList();
      document.getElementById("tagPageInfo").textContent =
        `Page ${data.page} / ${tagTotalPage}`;
    });
}
function renderTagList() {
  const list = document.getElementById("tagList");
  list.innerHTML = "";
  tagListData.forEach((tag) => {
    const chk = selectedTags.some((t) => t.id === tag.id);
    const label = document.createElement("label");
    label.className =
      "flex items-center gap-2 px-3 py-2 rounded-lg border border-transparent hover:bg-blue-50 cursor-pointer" +
      (chk ? " bg-blue-100 border-blue-300 font-semibold" : " bg-white");
    const cb = document.createElement("input");
    cb.type = "checkbox";
    cb.className = "accent-blue-500 mr-2 w-4 h-4";
    cb.checked = chk;
    cb.onchange = () => {
      cb.checked
        ? selectedTags.push({ id: tag.id, name: tag.name })
        : (selectedTags = selectedTags.filter((t) => t.id !== tag.id));
      renderTagList();
      renderSelectedTags();
      resetAndLoadNews();
    };
    const chip = document.createElement("span");
    chip.className =
      "inline-flex items-center px-3 py-0.5 rounded-2xl border text-blue-700 bg-blue-50 border-blue-200 text-xs font-medium leading-tight" +
      (chk ? " bg-blue-600 text-white border-blue-500" : "");
    chip.textContent = tag.name;
    const cnt = document.createElement("span");
    cnt.className = "ml-auto text-gray-400 text-xs";
    cnt.textContent = tag.news_count;
    label.append(cb, chip, cnt);
    list.appendChild(label);
  });
  renderSelectedTags();
}
function renderSelectedTags() {
  const wrap = document.getElementById("selectedTags");
  wrap.innerHTML = "";
  selectedTags.forEach((tag) => {
    const span = document.createElement("span");
    span.className =
      "inline-flex items-center px-3 py-0.5 rounded-2xl border text-white bg-blue-600 border-blue-500 text-xs font-medium mr-1 mb-1";
    span.textContent = tag.name;
    const x = document.createElement("span");
    x.className = "cursor-pointer ml-1 text-base font-bold";
    x.textContent = "×";
    x.onclick = () => {
      selectedTags = selectedTags.filter((t) => t.id !== tag.id);
      renderTagList();
      resetAndLoadNews();
    };
    span.appendChild(x);
    wrap.appendChild(span);
  });
}

/* ========= 뉴스 API ========= */
function resetAndLoadNews() {
  currentPage = 1;
  cardItems = [];
  isLoading = isEndOfList = false;
  document.getElementById("cardList").innerHTML = "";
  document.getElementById("newsCount").textContent = "";
  loadNewsPage();
}
function loadNewsPage() {
  if (isLoading || isEndOfList) return;
  isLoading = true;
  document.getElementById("loading").style.display = "";
  let url = `/api/whatsnews?limit=${PAGE_SIZE}&offset=${cardItems.length}`;
  if (selectedTags.length)
    url += `&tags=${selectedTags.map((t) => t.id).join(",")}`;
  if (newsSearchKeyword)
    url += `&search=${encodeURIComponent(newsSearchKeyword)}`;
  fetch(url)
    .then((r) => r.json())
    .then((data) => {
      const items = data.items || [];
      cardItems = cardItems.concat(items);
      totalCount = data.total || 0;
      fillCardList(cardItems);
      document.getElementById("newsCount").textContent = `전체 ${totalCount}건`;
      if (cardItems.length >= totalCount || !items.length) isEndOfList = true;
    })
    .finally(() => {
      isLoading = false;
      document.getElementById("loading").style.display = "none";
    });
}

/* ========= 카드 렌더 ========= */
function fillCardList(items) {
  const list = document.getElementById("cardList");
  list.innerHTML = "";
  items.forEach((it) => {
    const card = document.createElement("div");
    card.className =
      "bg-white border border-blue-100 rounded-2xl shadow-lg flex flex-col min-w-0 min-h-[200px] max-w-3xl mx-auto px-6 py-6";

    const title = document.createElement("div");
    title.className = "text-lg font-semibold mb-3 leading-tight break-all";
    if (it.source_url) {
      let u = it.source_url;
      if (!/^https?:\/\//.test(u))
        u = "https://aws.amazon.com" + (u.startsWith("/") ? u : "/" + u);
      const a = document.createElement("a");
      a.href = u;
      a.target = "_blank";
      a.textContent = it.title;
      a.className = "text-blue-700 hover:underline";
      title.appendChild(a);
    } else title.textContent = it.title;

    const date = document.createElement("div");
    date.className = "text-gray-500 text-sm mb-2";
    date.textContent = it.source_created_at
      ? it.source_created_at.slice(0, 10)
      : "";

    const tagWrap = document.createElement("div");
    tagWrap.className = "flex flex-wrap gap-2 mb-3";
    (it.tags || []).forEach((tg) => {
      const s = document.createElement("span");
      s.className =
        "inline-flex items-center px-3 py-0.5 rounded-2xl border text-blue-700 bg-blue-50 border-blue-200 text-xs font-medium leading-tight";
      s.textContent = tg.name;
      tagWrap.appendChild(s);
    });

    const body = document.createElement("div");
    body.className = "text-base text-gray-900 mt-1";
    if (it.content) body.innerHTML = it.content;

    card.append(title, date, tagWrap, body);
    list.appendChild(card);
  });
}

/* ========= 무한 스크롤 ========= */
function onMainScroll(e) {
  const el = e.target;
  if (isLoading || isEndOfList) return;
  if (el.scrollHeight - el.scrollTop - el.clientHeight < 80) loadNewsPage();
}
